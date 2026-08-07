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

		// DependsOn returns the names of modules that must be registered,
		// and whose RegisterServices/InitializeFromModules hooks must run,
		// before this module's own. composer.Initialize topologically sorts
		// on these names (ties broken by module name, so the order is
		// stable across runs) and fails with a dependency-cycle error
		// instead of running modules in an arbitrary order.
		DependsOn() []string
	}
)
