package nomoyu

import (
	"context"

	"github.com/nomoyu/go-gin-framework/pkg/logger"
	"github.com/nomoyu/go-gin-framework/pkg/scheduler"
)

// SchedulerOption 控制框架内置的定时任务调度。
type SchedulerOption struct {
	Instances []*scheduler.Scheduler
	Tasks     []scheduler.Task
}

// WithScheduler 直接指定调度器实例（可自定义时区等，支持多次调用追加多个实例）。
func (a *App) WithScheduler(schedulers ...*scheduler.Scheduler) *App {
	if a.scheduler == nil {
		a.scheduler = &SchedulerOption{}
	}
	for _, s := range schedulers {
		if s != nil {
			a.scheduler.Instances = append(a.scheduler.Instances, s)
		}
	}
	return a
}

// WithCronTasks 快速注册定时任务，未显式设置调度器时使用默认配置。
func (a *App) WithCronTasks(tasks ...scheduler.Task) *App {
	if a.scheduler == nil {
		a.scheduler = &SchedulerOption{}
	}
	a.scheduler.Tasks = append(a.scheduler.Tasks, tasks...)
	return a
}

func (a *App) startScheduler() {
	if a.scheduler == nil {
		return
	}

	// 至少保留一个调度器，用于挂载框架内配置的定时任务。
	if len(a.scheduler.Instances) == 0 || a.scheduler.Instances[0] == nil {
		a.scheduler.Instances = append([]*scheduler.Scheduler{scheduler.New()}, a.scheduler.Instances...)
	}

	// 框架层注册的任务挂到第一个调度器，其余调度器按用户预置任务启动。
	if first := a.scheduler.Instances[0]; first != nil {
		for _, task := range a.scheduler.Tasks {
			if _, err := first.AddTask(task); err != nil {
				logger.Errorf("register cron task %s failed: %v", task.Name, err)
			}
		}
	}

	for _, instance := range a.scheduler.Instances {
		if instance == nil {
			continue
		}
		instance.Start()
	}

	a.OnShutdown(func(ctx context.Context) error {
		for _, instance := range a.scheduler.Instances {
			if instance == nil {
				continue
			}
			if err := instance.Stop(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
