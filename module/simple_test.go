package module

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ Module = (*mockModule)(nil)

type mockModule struct {
	name string
}

func (m *mockModule) RegisterServices() func(do.Injector) {
	return nil
}

func (m *mockModule) DependsOn() []string {
	return nil
}

func (m *mockModule) Name() string {
	return m.name
}

func TestNewSimpleRegistry(t *testing.T) {
	registry := NewSimpleRegistry()
	assert.NotNil(t, registry)
	assert.IsType(t, &simpleRegistry{}, registry)
}

func TestSimpleRegistry_AddModule(t *testing.T) {
	registry := NewSimpleRegistry()
	module1 := &mockModule{name: "module1"}

	assert.NoError(t, registry.AddModule(module1))
	assert.True(t, registry.HasModule("module1"))

	assert.Error(t, registry.AddModule(module1))
}

func TestSimpleRegistry_MustAddModule(t *testing.T) {
	registry := NewSimpleRegistry()
	module1 := &mockModule{name: "module1"}

	registry.MustAddModule(module1)
	assert.True(t, registry.HasModule("module1"))

	assert.Panics(t, func() {
		registry.MustAddModule(module1)
	})
}

func TestSimpleRegistry_AddModules(t *testing.T) {
	registry := NewSimpleRegistry()
	module1 := &mockModule{name: "module1"}
	module2 := &mockModule{name: "module2"}
	module3 := &mockModule{name: "module3"}

	require.NoError(t, registry.AddModules(module1, module2))
	assert.True(t, registry.HasModule("module1"))
	assert.True(t, registry.HasModule("module2"))

	// Stops at the first duplicate: module3 after it is not added.
	err := registry.AddModules(module1, module3)
	require.ErrorIs(t, err, ErrModuleAlreadyRegistered)
	assert.False(t, registry.HasModule("module3"))
}

func TestSimpleRegistry_MustAddModules(t *testing.T) {
	registry := NewSimpleRegistry()
	module1 := &mockModule{name: "module1"}
	module2 := &mockModule{name: "module2"}

	registry.MustAddModules(module1, module2)
	assert.True(t, registry.HasModule("module1"))
	assert.True(t, registry.HasModule("module2"))

	assert.Panics(t, func() {
		registry.MustAddModules(module2)
	})
}

func TestSimpleRegistry_ConcurrentAddModule(t *testing.T) {
	const workers = 16

	registry := NewSimpleRegistry()

	var (
		wg        sync.WaitGroup
		successes atomic.Int32
	)

	for range workers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if registry.AddModule(&mockModule{name: "shared"}) == nil {
				successes.Add(1)
			}

			registry.HasModule("shared")
			registry.GetModule("shared")
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(1), successes.Load(), "exactly one AddModule must win")
	assert.Len(t, registry.GetAllModules(), 1)
}

func TestSimpleRegistry_HasModule(t *testing.T) {
	registry := NewSimpleRegistry()
	module1 := &mockModule{name: "module1"}

	assert.False(t, registry.HasModule("module1"))
	registry.MustAddModule(module1)
	assert.True(t, registry.HasModule("module1"))
}

func TestSimpleRegistry_GetModule(t *testing.T) {
	registry := NewSimpleRegistry()
	module1 := &mockModule{name: "module1"}

	registry.MustAddModule(module1)

	result := registry.GetModule("module1")
	assert.Equal(t, module1, result)

	nonExistent := registry.GetModule("non-existent")
	assert.Nil(t, nonExistent)
}

func TestSimpleRegistry_GetAllModules(t *testing.T) {
	registry := NewSimpleRegistry()
	module1 := &mockModule{name: "module1"}
	module2 := &mockModule{name: "module2"}

	registry.MustAddModule(module1)
	registry.MustAddModule(module2)

	modules := registry.GetAllModules()
	assert.Len(t, modules, 2)
	assert.Contains(t, modules, module1)
	assert.Contains(t, modules, module2)
}
