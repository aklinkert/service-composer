package servicecomposer_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/samber/do/v2"

	servicecomposer "src/github.com/aklinkert/go-service-composer"
)

type ServiceConfig struct {
	ListenAPI string
}

func (s *ServiceConfig) GetListenAPI() string {
	return s.ListenAPI
}

type exampleModule struct {
	config *ServiceConfig
	logger *slog.Logger
}

func newExampleModule(config *ServiceConfig, logger *slog.Logger) *exampleModule {
	return &exampleModule{
		config: config,
		logger: logger,
	}
}

func (m *exampleModule) Name() string {
	return "example"
}

func (m *exampleModule) RegisterServices() func(_ do.Injector) {
	return do.Package()
}

func (m *exampleModule) DependsOn() []string {
	return []string{}
}

func (m *exampleModule) Start(_ context.Context, _ do.Injector) error {
	m.logger.Info("Started HTTP Server", "listen", m.config.ListenAPI)

	return nil
}

func getTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,

		// omit time for test output consistency
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.Attr{}
			}

			return a
		},
	}))
}

func ExampleNew() {
	logger := getTestLogger()

	// provide your own config while fulfilling the required module's interfaces
	config := &ServiceConfig{
		ListenAPI: ":8080",
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	eco := servicecomposer.New(ctx, logger, "example_service")

	// add your modules, we're adding an example module here
	eco.MustAddModule(newExampleModule(config, logger))

	go func() {
		time.Sleep(10 * time.Millisecond)

		cancelCtx()
	}()

	if err := eco.Start(); err != nil {
		fmt.Println(err)
	}

	// Output:
	// level=DEBUG msg="starting module" module=example
	// level=INFO msg="Started HTTP Server" listen=:8080
}
