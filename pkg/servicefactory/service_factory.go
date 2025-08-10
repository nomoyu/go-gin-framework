package servicefactory

import "sync"

// ServiceFactory 定义对象创建器
type ServiceFactory struct {
	creators  map[string]func() any
	instances map[string]any
	mu        sync.Mutex
}

// global singleton
var factory = &ServiceFactory{
	creators:  make(map[string]func() any),
	instances: make(map[string]any),
}

// Register 注册一个对象创建器
func Register(name string, creator func() any) {
	factory.mu.Lock()
	defer factory.mu.Unlock()
	factory.creators[name] = creator
}

// GetInstance 获取或创建单例对象
func GetInstance(name string) (any, error) {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	if instance, ok := factory.instances[name]; ok {
		return instance, nil
	}

	creator, ok := factory.creators[name]
	if !ok {
		return nil, errNotRegistered(name)
	}

	instance := creator()
	factory.instances[name] = instance
	return instance, nil
}

// ErrNotRegistered 自定义错误
type errNotRegistered string

func (e errNotRegistered) Error() string {
	return "object servicefactory: not registered - " + string(e)
}
