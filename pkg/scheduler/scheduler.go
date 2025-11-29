package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nomoyu/go-gin-framework/pkg/logger"
)

// Option defines a functional option to configure scheduler behavior.
type Option func(*options)

type options struct {
	location *time.Location
}

// Task 定义了一个可运行的定时任务。
type Task struct {
	Name string
	Spec string
	Job  func(ctx context.Context) error
}

// Scheduler 封装定时任务调度器，统一日志与生命周期。
type Scheduler struct {
	tasks   []scheduledTask
	started bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.Mutex
	options options
}

type scheduledTask struct {
	Task
	runner taskRunner
}

type taskRunner func(ctx context.Context, run func())

// New 创建一个 Scheduler，可选自定义时区（默认 time.Local）。
func New(opts ...Option) *Scheduler {
	s := &Scheduler{
		options: options{location: time.Local},
	}
	for _, opt := range opts {
		opt(&s.options)
	}
	return s
}

// WithLocation 设置定时任务的时区，默认使用 time.Local。
func WithLocation(loc *time.Location) Option {
	return func(o *options) {
		if loc != nil {
			o.location = loc
		}
	}
}

// AddTask 注册一个定时任务，支持以下 Spec 形式：
//  1. 纯 duration（如 "5m"），等价于 @every 5m
//  2. "@every <duration>"，按固定间隔循环
//  3. "@daily" 或 "@daily HH:MM(:SS)", 每天指定时间
//  4. "@hourly" 或 "@hourly MM(:SS)", 每小时指定分钟
func (s *Scheduler) AddTask(task Task) (int, error) {
	if task.Job == nil {
		return 0, errors.New("task job cannot be nil")
	}
	if strings.TrimSpace(task.Spec) == "" {
		return 0, errors.New("task spec cannot be empty")
	}

	runner, err := buildRunner(task.Spec, s.options.location)
	if err != nil {
		return 0, err
	}

	name := task.Name
	if name == "" {
		name = task.Spec
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, scheduledTask{Task: Task{Name: name, Spec: task.Spec, Job: task.Job}, runner: runner})
	return len(s.tasks), nil
}

// Start 启动调度器（幂等）。
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.started = true

	for _, task := range s.tasks {
		s.wg.Add(1)
		go s.runTask(ctx, task)
	}
	s.mu.Unlock()
	logger.Infof("cron scheduler started")
}

// Stop 停止调度器，等待正在运行的任务结束或上下文超时。
func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = false
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Infof("cron scheduler stopped")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("stop scheduler timeout: %w", ctx.Err())
	}
}

func (s *Scheduler) runTask(ctx context.Context, task scheduledTask) {
	defer s.wg.Done()

	safeRun := func() {
		start := time.Now()
		logger.Infof("cron task start: %s (%s)", task.Name, task.Spec)
		if err := task.Job(context.Background()); err != nil {
			logger.Errorf("cron task failed: %s (%s), err=%v", task.Name, task.Spec, err)
		} else {
			logger.Infof("cron task done: %s (%s) in %s", task.Name, task.Spec, time.Since(start))
		}
	}

	task.runner(ctx, safeRun)
}

func buildRunner(spec string, loc *time.Location) (taskRunner, error) {
	if loc == nil {
		loc = time.Local
	}

	trimmed := strings.TrimSpace(spec)
	if strings.HasPrefix(trimmed, "@every") {
		parts := strings.Fields(trimmed)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid @every spec: %s", spec)
		}
		interval, err := time.ParseDuration(parts[1])
		if err != nil || interval <= 0 {
			return nil, fmt.Errorf("invalid @every duration: %s", parts[1])
		}
		return everyRunner(interval), nil
	}

	if strings.HasPrefix(trimmed, "@daily") {
		at, err := parseTimeOfDay(trimmed, "@daily", 0, 0, 0)
		if err != nil {
			return nil, err
		}
		return dailyRunner(loc, at), nil
	}

	if strings.HasPrefix(trimmed, "@hourly") {
		minute, second, err := parseMinuteSecond(trimmed, "@hourly", 0, 0)
		if err != nil {
			return nil, err
		}
		return hourlyRunner(minute, second), nil
	}

	if dur, err := time.ParseDuration(trimmed); err == nil {
		if dur <= 0 {
			return nil, fmt.Errorf("duration must be positive: %s", spec)
		}
		return everyRunner(dur), nil
	}

	return nil, fmt.Errorf("unsupported spec: %s", spec)
}

func everyRunner(interval time.Duration) taskRunner {
	return func(ctx context.Context, run func()) {
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				run()
				timer.Reset(interval)
			}
		}
	}
}

func dailyRunner(loc *time.Location, at time.Duration) taskRunner {
	return func(ctx context.Context, run func()) {
		for {
			now := time.Now().In(loc)
			next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Add(at)
			if !next.After(now) {
				next = next.Add(24 * time.Hour)
			}
			wait := time.Until(next)
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				run()
			}
		}
	}
}

func hourlyRunner(minute, second int) taskRunner {
	return func(ctx context.Context, run func()) {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), minute, second, 0, now.Location())
			if !next.After(now) {
				next = next.Add(time.Hour)
			}
			wait := time.Until(next)
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				run()
			}
		}
	}
}

func parseTimeOfDay(spec, prefix string, defaultHour, defaultMinute, defaultSecond int) (time.Duration, error) {
	fields := strings.Fields(spec)
	if len(fields) == 1 {
		return time.Duration(defaultHour)*time.Hour + time.Duration(defaultMinute)*time.Minute + time.Duration(defaultSecond)*time.Second, nil
	}
	if len(fields) != 2 {
		return 0, fmt.Errorf("invalid %s spec: %s", prefix, spec)
	}

	parts := strings.Split(fields[1], ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, fmt.Errorf("invalid %s time: %s", prefix, fields[1])
	}

	hour, err := parseClockValue(parts[0], 0, 23)
	if err != nil {
		return 0, fmt.Errorf("invalid %s hour: %w", prefix, err)
	}
	minute, err := parseClockValue(parts[1], 0, 59)
	if err != nil {
		return 0, fmt.Errorf("invalid %s minute: %w", prefix, err)
	}
	second := 0
	if len(parts) == 3 {
		second, err = parseClockValue(parts[2], 0, 59)
		if err != nil {
			return 0, fmt.Errorf("invalid %s second: %w", prefix, err)
		}
	}

	return time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute + time.Duration(second)*time.Second, nil
}

func parseMinuteSecond(spec, prefix string, defaultMinute, defaultSecond int) (int, int, error) {
	fields := strings.Fields(spec)
	if len(fields) == 1 {
		return defaultMinute, defaultSecond, nil
	}
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("invalid %s spec: %s", prefix, spec)
	}

	parts := strings.Split(fields[1], ":")
	if len(parts) < 1 || len(parts) > 2 {
		return 0, 0, fmt.Errorf("invalid %s time: %s", prefix, fields[1])
	}

	minute, err := parseClockValue(parts[0], 0, 59)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid %s minute: %w", prefix, err)
	}
	second := defaultSecond
	if len(parts) == 2 {
		second, err = parseClockValue(parts[1], 0, 59)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid %s second: %w", prefix, err)
		}
	}

	return minute, second, nil
}

func parseClockValue(value string, min, max int) (int, error) {
	iv, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if iv < min || iv > max {
		return 0, fmt.Errorf("value %d out of range [%d,%d]", iv, min, max)
	}
	return iv, nil
}
