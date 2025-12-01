package module

import (
	"fmt"
	"sync"
)

var (
	ErrModuleAlreadyRegistered = fmt.Errorf("module already registered")
)

type simpleRegistry struct {
	modulesLock sync.RWMutex
	modules     map[string]Module
}

func NewSimpleRegistry() Registry {
	return &simpleRegistry{
		modules: map[string]Module{},
	}
}

func (r *simpleRegistry) MustAddModules(modules ...Module) {
	for _, module := range modules {
		r.MustAddModule(module)
	}
}

func (r *simpleRegistry) AddModules(modules ...Module) error {
	for _, module := range modules {
		if err := r.AddModule(module); err != nil {
			return err
		}
	}

	return nil
}

func (r *simpleRegistry) MustAddModule(module Module) {
	if err := r.AddModule(module); err != nil {
		panic(err)
	}
}

func (r *simpleRegistry) AddModule(module Module) error {
	if r.HasModule(module.Name()) {
		return fmt.Errorf("%w: %s", ErrModuleAlreadyRegistered, module.Name())
	}

	r.modulesLock.Lock()
	r.modules[module.Name()] = module
	r.modulesLock.Unlock()

	return nil
}

func (r *simpleRegistry) HasModule(name string) bool {
	for _, module := range r.modules {
		if module.Name() == name {
			return true
		}
	}

	return false
}

func (r *simpleRegistry) GetModule(name string) Module {
	if !r.HasModule(name) {
		return nil
	}

	r.modulesLock.RLock()
	module := r.modules[name]
	r.modulesLock.RUnlock()

	return module
}

func (r *simpleRegistry) GetAllModules() []Module {
	modules := make([]Module, 0)

	r.modulesLock.RLock()

	for _, module := range r.modules {
		modules = append(modules, module)
	}

	r.modulesLock.RUnlock()

	return modules
}
