package servicecomposer

import (
	"context"

	"github.com/samber/do/v2"

	"github.com/aklinkert/service-composer/module"
)

type (
	Composer interface {
		context.Context
		module.Registry

		Injector() do.Injector

		// Initialize runs every registered module's RegisterServices, then
		// every module's InitializeFromModules, each pass ordered so a
		// module's DependsOn() dependencies run before it does (see
		// module.Module.DependsOn). It returns a dependency-cycle error
		// instead of running modules in an arbitrary order if the graph
		// can't be sorted. All modules must be registered before calling
		// Initialize; Start rejects any registered afterwards.
		Initialize() error
		Start() error
	}

	Logger interface {
		Info(message string, args ...any)
		Debug(message string, args ...any)
	}

	StartableModule interface {
		Start(ctx context.Context, injector do.Injector) error
	}

	FromModulesInitializerModule interface {
		InitializeFromModules(ctx context.Context, injector do.Injector, modules []module.Module) error
	}

	Option interface {
		apply(*composer)
	}
)
