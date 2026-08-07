package servicecomposer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aklinkert/service-composer/module"
)

// visitState tracks a module's progress through the cycle-detection DFS in
// newCycleError.
type visitState int

const (
	visitStateUnvisited visitState = iota
	visitStateVisiting
	visitStateVisited
)

// topologicalSort orders modules so that every module's dependencies (per
// DependsOn()) precede the module itself. Ties among modules that have no
// remaining unresolved dependency are broken by module name, so the result
// is stable and reproducible across calls rather than depending on Go's
// randomised map iteration order.
//
// topologicalSort assumes every name returned by DependsOn() refers to a
// module present in modules; callers are expected to validate that
// separately (composer.Initialize does, via its own presence check) so a
// stall here is reported as a dependency cycle rather than a missing
// module.
func topologicalSort(modules []module.Module) ([]module.Module, error) {
	byName := make(map[string]module.Module, len(modules))
	// depsByName snapshots DependsOn() once per module so the rest of this
	// function - including the cycle-path DFS below - never calls it again.
	// Reading it more than once would let a module whose DependsOn() isn't
	// stable across calls silently desynchronise the dependency count from
	// the edges built from it.
	depsByName := make(map[string][]string, len(modules))
	remaining := make(map[string]int, len(modules))
	dependents := make(map[string][]string, len(modules))

	for _, mod := range modules {
		byName[mod.Name()] = mod
		deps := append([]string(nil), mod.DependsOn()...)
		depsByName[mod.Name()] = deps
		remaining[mod.Name()] = len(deps)

		for _, dep := range deps {
			dependents[dep] = append(dependents[dep], mod.Name())
		}
	}

	ready := make([]string, 0, len(modules))

	for _, mod := range modules {
		if remaining[mod.Name()] == 0 {
			ready = append(ready, mod.Name())
		}
	}

	sort.Strings(ready)

	ordered := make([]module.Module, 0, len(modules))

	for len(ready) > 0 {
		name := ready[0]
		ready = ready[1:]

		ordered = append(ordered, byName[name])

		freed := releaseDependents(dependents[name], remaining)
		if len(freed) == 0 {
			continue
		}

		ready = append(ready, freed...)
		sort.Strings(ready)
	}

	if len(ordered) == len(modules) {
		return ordered, nil
	}

	return nil, newCycleError(depsByName, remaining)
}

// releaseDependents decrements the remaining-dependency count of each
// dependent and returns those that reached zero, i.e. became ready to be
// scheduled next.
func releaseDependents(dependents []string, remaining map[string]int) []string {
	freed := make([]string, 0, len(dependents))

	for _, dependent := range dependents {
		remaining[dependent]--

		if remaining[dependent] == 0 {
			freed = append(freed, dependent)
		}
	}

	return freed
}

// newCycleError runs a deterministic DFS, in module-name order, over the
// modules that never reached zero remaining dependencies and reports the
// first cycle it finds as e.g. "a -> b -> c -> a".
func newCycleError(depsByName map[string][]string, remaining map[string]int) error {
	stuck := make([]string, 0, len(remaining))

	for name, count := range remaining {
		if count > 0 {
			stuck = append(stuck, name)
		}
	}

	sort.Strings(stuck)

	state := make(map[string]visitState, len(stuck))

	for _, name := range stuck {
		if state[name] != visitStateUnvisited {
			continue
		}

		if cycle := findCycle(name, depsByName, remaining, state, nil); cycle != nil {
			return fmt.Errorf("%w: %s", ErrDependencyCycle, strings.Join(cycle, " -> "))
		}
	}

	// Reachable only if a module in modules depends on a name that is
	// itself never registered: topologicalSort assumes callers already
	// validated presence (composer.Initialize does), so this stall has no
	// discoverable cycle to report - name it explicitly rather than
	// mislabeling a missing module as one.
	return fmt.Errorf("%w: modules %s never resolved their dependencies (check for a DependsOn() name with no matching registered module)", ErrDependencyCycle, strings.Join(stuck, ", "))
}

// findCycle walks the dependency graph, restricted to modules still stuck
// (remaining[name] > 0), depth-first in dependency-name order, and returns
// the first cycle it encounters as an ordered slice looping back to its own
// start, e.g. path=[a,b], dep b depends on a => [a,b,a].
func findCycle(name string, depsByName map[string][]string, remaining map[string]int, state map[string]visitState, path []string) []string {
	state[name] = visitStateVisiting
	path = append(path, name)

	// depsByName[name] is already our own private copy (see
	// topologicalSort), so sorting it in place is safe here.
	deps := depsByName[name]
	sort.Strings(deps)

	for _, dep := range deps {
		if remaining[dep] == 0 {
			continue
		}

		if state[dep] == visitStateVisiting {
			return closeCycle(path, dep)
		}

		if state[dep] == visitStateUnvisited {
			if cycle := findCycle(dep, depsByName, remaining, state, path); cycle != nil {
				return cycle
			}
		}
	}

	state[name] = visitStateVisited

	return nil
}

// closeCycle returns the suffix of path starting at start, followed by
// start again, e.g. path=[a,b,c], start=b => [b,c,b].
func closeCycle(path []string, start string) []string {
	for i, name := range path {
		if name == start {
			return append(append([]string{}, path[i:]...), start)
		}
	}

	return append(append([]string{}, path...), start)
}
