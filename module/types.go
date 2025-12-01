package module

import (
	"github.com/samber/do/v2"
)

type (
	Registry interface {
		AddModules(modules ...Module) error
		MustAddModules(modules ...Module)

		AddModule(module Module) error
		MustAddModule(module Module)

		HasModule(name string) bool

		GetModule(name string) Module
		GetAllModules() []Module
	}

	Module interface {
		Name() string
		RegisterServices() func(do.Injector)
		DependsOn() []string
	}
)
