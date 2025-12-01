package module

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
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

	assert.NoError(t, registry.AddModule(module1))
	assert.True(t, registry.HasModule("module1"))
	assert.NoError(t, registry.AddModule(module2))
	assert.True(t, registry.HasModule("module2"))

	assert.Error(t, registry.AddModule(module1))
	assert.Error(t, registry.AddModule(module2))
}

func TestSimpleRegistry_MustAddModules(t *testing.T) {
	registry := NewSimpleRegistry()
	module1 := &mockModule{name: "module1"}
	module2 := &mockModule{name: "module2"}

	registry.MustAddModule(module1)
	assert.True(t, registry.HasModule("module1"))

	registry.MustAddModule(module2)
	assert.True(t, registry.HasModule("module2"))

	assert.Panics(t, func() {
		registry.MustAddModule(module1)
	})
	assert.Panics(t, func() {
		registry.MustAddModule(module2)
	})
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
