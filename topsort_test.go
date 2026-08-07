package servicecomposer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aklinkert/service-composer/module"
)

func namesOf(modules []module.Module) []string {
	names := make([]string, 0, len(modules))

	for _, mod := range modules {
		names = append(names, mod.Name())
	}

	return names
}

// TestTopologicalSort_OrdersDependenciesBeforeDependents pins down an
// unambiguous graph (a has no deps; b and c each depend only on a and are
// independent of each other; d depends on both) and asserts the sort always
// produces the same order regardless of the order modules are handed in.
// topologicalSort has no map-iteration-dependent behaviour on this
// success path (byName/remaining/dependents are only ever looked up by
// key, never ranged, before ordered is complete), so a single pass per
// input order is enough to pin its determinism; it's
// TestComposer_Initialize_DeterministicOrderAcrossRuns, driven through the
// real map-backed module.Registry, that repeats across many runs to catch
// anything that does depend on Go's per-range-call randomised map order.
func TestTopologicalSort_OrdersDependenciesBeforeDependents(t *testing.T) {
	expected := []string{"a", "b", "c", "d"}

	inputOrders := [][]string{
		{"a", "b", "c", "d"},
		{"d", "c", "b", "a"},
		{"b", "d", "a", "c"},
		{"c", "a", "d", "b"},
		{"d", "a", "c", "b"},
	}

	for _, order := range inputOrders {
		modules := buildDiamondGraph(order)

		ordered, err := topologicalSort(modules)
		require.NoError(t, err)
		assert.Equal(t, expected, namesOf(ordered))
	}
}

// buildDiamondGraph returns the a/b/c/d modules used above, added to the
// returned slice in the given name order.
func buildDiamondGraph(order []string) []module.Module {
	byName := map[string]module.Module{
		"a": newTestModule("a", nil, ""),
		"b": newTestModule("b", []string{"a"}, ""),
		"c": newTestModule("c", []string{"a"}, ""),
		"d": newTestModule("d", []string{"b", "c"}, ""),
	}

	modules := make([]module.Module, 0, len(order))

	for _, name := range order {
		modules = append(modules, byName[name])
	}

	return modules
}

func TestTopologicalSort_DependencyCycle(t *testing.T) {
	a := newTestModule("a", []string{"b"}, "")
	b := newTestModule("b", []string{"c"}, "")
	c := newTestModule("c", []string{"a"}, "")

	_, err := topologicalSort([]module.Module{c, a, b})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDependencyCycle)
	assert.True(t, strings.Contains(err.Error(), "a -> b -> c -> a"), "error should name the cycle, got: %s", err.Error())
}

func TestTopologicalSort_SelfDependency(t *testing.T) {
	a := newTestModule("a", []string{"a"}, "")

	_, err := topologicalSort([]module.Module{a})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDependencyCycle)
	assert.Contains(t, err.Error(), "a -> a")
}

func TestTopologicalSort_CycleAmongstOtherwiseResolvableModules(t *testing.T) {
	// root has no deps and resolves fine; x and y form a cycle that must
	// still be reported even though root is not involved.
	root := newTestModule("root", nil, "")
	x := newTestModule("x", []string{"y"}, "")
	y := newTestModule("y", []string{"x"}, "")

	_, err := topologicalSort([]module.Module{y, root, x})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDependencyCycle)
	assert.Contains(t, err.Error(), "x -> y -> x")
}

func TestTopologicalSort_NoModules(t *testing.T) {
	ordered, err := topologicalSort(nil)
	require.NoError(t, err)
	assert.Empty(t, ordered)
}

// TestTopologicalSort_MissingDependency pins down the documented, currently
// untested fallback behaviour when topologicalSort's presence precondition
// is violated: a's dependency "ghost" is never registered. composer.
// Initialize never lets this reach topologicalSort (it validates presence
// itself first, with its own ErrModuleNotFound message), but topologicalSort
// is also called directly by these tests, so its own behaviour on a stall
// with no discoverable cycle needs to be pinned rather than left as dead,
// unasserted code.
func TestTopologicalSort_MissingDependency(t *testing.T) {
	a := newTestModule("a", []string{"ghost"}, "")

	_, err := topologicalSort([]module.Module{a})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDependencyCycle)
	assert.Contains(t, err.Error(), "a")
}

// TestTopologicalSort_TwoDisjointCycles pins the "report the
// alphabetically-first cycle" contract: x/y and m/n each form their own
// cycle, unrelated to each other, and newCycleError visits stuck modules in
// name order, so it must report the m/n cycle first.
func TestTopologicalSort_TwoDisjointCycles(t *testing.T) {
	x := newTestModule("x", []string{"y"}, "")
	y := newTestModule("y", []string{"x"}, "")
	m := newTestModule("m", []string{"n"}, "")
	n := newTestModule("n", []string{"m"}, "")

	_, err := topologicalSort([]module.Module{x, y, m, n})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDependencyCycle)
	assert.Contains(t, err.Error(), "m -> n -> m")
	assert.NotContains(t, err.Error(), "x -> y -> x")
}
