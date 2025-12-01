package servicecomposer

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"

	"src/github.com/aklinkert/go-service-composer/module"
)

var (
	_ module.Module                = (*testModule[string])(nil)
	_ StartableModule              = (*testModule[string])(nil)
	_ FromModulesInitializerModule = (*testModule[string])(nil)
)

type testModule[T any] struct {
	name      string
	dependsOn []string
	value     T

	initialized bool
	started     bool
}

func newTestModule[T any](name string, dependsOn []string, value T) *testModule[T] {
	return &testModule[T]{
		name:      name,
		dependsOn: dependsOn,
		value:     value,
	}
}

func (m *testModule[T]) Name() string {
	return m.name
}

func (m *testModule[T]) RegisterServices() func(_ do.Injector) {
	return do.Package(
		do.LazyNamed(fmt.Sprintf("%v.value", m.Name()), func(_ do.Injector) (T, error) {
			return m.value, nil
		}),
	)
}

func (m *testModule[T]) DependsOn() []string {
	return m.dependsOn
}

func (m *testModule[T]) InitializeFromModules(_ context.Context, _ do.Injector, _ []module.Module) error {
	m.initialized = true

	return nil
}

func (m *testModule[T]) Start(_ context.Context, _ do.Injector) error {
	m.started = true

	return nil
}

type testServer struct {
	started bool
	stopped bool
}

func (s *testServer) Start() error {
	s.started = true

	return nil
}

func (s *testServer) Shutdown() error {
	s.stopped = true

	return nil
}
