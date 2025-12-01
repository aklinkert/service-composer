package servicecomposer

import (
	"context"

	"github.com/samber/do/v2"

	"src/github.com/aklinkert/service-composer/module"
)

type (
	Composer interface {
		context.Context
		module.Registry

		Injector() do.Injector
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
