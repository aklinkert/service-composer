package servicecomposer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/do/v2"

	"src/github.com/aklinkert/service-composer/module"
)

type composer struct {
	context.Context
	module.Registry

	serviceName string

	injector *do.RootScope
	logger   Logger

	initialized bool

	// options
	moduleInitTimeout time.Duration
}

func New(ctx context.Context, logger Logger, serviceName string, options ...Option) Composer {
	c := &composer{
		Context:     ctx,
		Registry:    module.NewSimpleRegistry(),
		serviceName: serviceName,
		injector:    do.New(),

		moduleInitTimeout: DefaultModuleInitTimeout,

		logger: logger,
	}

	do.ProvideNamedValue(c.injector, "service_name", c.serviceName)
	do.ProvideValue(c.injector, c.Context)

	for _, opt := range options {
		opt.apply(c)
	}

	return c
}

func (c *composer) Injector() do.Injector {
	return c.injector
}

func (c *composer) Initialize() error {
	if c.initialized {
		return nil
	}

	for _, mod := range c.GetAllModules() {
		dependsOn := mod.DependsOn()

		for _, dep := range dependsOn {
			if !c.HasModule(dep) {
				return fmt.Errorf("%w: module %s depends on %s, but it is not registered", ErrModuleNotFound, mod.Name(), dep)
			}
		}
	}

	allModules := c.GetAllModules()

	for _, mod := range allModules {
		if register := mod.RegisterServices(); register != nil {
			register(c.injector)
		}
	}

	for _, mod := range allModules {
		initializer, ok := mod.(FromModulesInitializerModule)
		if !ok {
			continue
		}

		c.logger.Debug("initializing module", slog.String("module", mod.Name()))

		moduleInitCtx, cancel := context.WithTimeout(c.Context, c.moduleInitTimeout)
		err := initializer.InitializeFromModules(moduleInitCtx, c.injector, allModules)

		cancel()

		if err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", mod.Name(), err)
		}
	}

	c.initialized = true

	return nil
}

func (c *composer) Start() error {
	if err := c.Initialize(); err != nil {
		return err
	}

	for _, mod := range c.GetAllModules() {
		starter, ok := mod.(StartableModule)
		if !ok {
			continue
		}

		c.logger.Debug("starting module", slog.String("module", mod.Name()))

		if err := starter.Start(c.Context, c.injector); err != nil {
			return fmt.Errorf("%w %s: %w", ErrFailedStart, mod.Name(), err)
		}
	}

	<-c.Done()

	if report := c.injector.Shutdown(); !report.Succeed {
		return report
	}

	return nil
}
