package servicecomposer

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aklinkert/service-composer/module"
)

// orderRecordingModule appends a "<name>-<pass>" entry to a shared, pointed-to
// slice from each lifecycle hook it implements, so tests can assert the
// exact sequence Initialize/Start ran the hooks in.
type orderRecordingModule struct {
	name      string
	dependsOn []string
	order     *[]string
}

func newOrderRecordingModule(name string, dependsOn []string, order *[]string) *orderRecordingModule {
	return &orderRecordingModule{name: name, dependsOn: dependsOn, order: order}
}

func (m *orderRecordingModule) Name() string {
	return m.name
}

func (m *orderRecordingModule) DependsOn() []string {
	return m.dependsOn
}

func (m *orderRecordingModule) RegisterServices() func(do.Injector) {
	return func(_ do.Injector) {
		*m.order = append(*m.order, m.name+"-register")
	}
}

func (m *orderRecordingModule) InitializeFromModules(_ context.Context, _ do.Injector, _ []module.Module) error {
	*m.order = append(*m.order, m.name+"-init")

	return nil
}

func (m *orderRecordingModule) Start(_ context.Context, _ do.Injector) error {
	*m.order = append(*m.order, m.name+"-start")

	return nil
}

func newTestComposer(ctx context.Context) *composer {
	return New(ctx, slog.New(slog.NewTextHandler(io.Discard, nil)), "test").(*composer)
}

func filterSuffix(entries []string, suffix string) []string {
	filtered := make([]string, 0, len(entries))

	for _, entry := range entries {
		if strings.HasSuffix(entry, suffix) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// TestComposer_Initialize_DeterministicOrderAcrossRuns gives high-confidence
// regression coverage through the real registry (GetAllModules is backed by
// a Go map, whose iteration order is randomised per range call, not just
// per process) rather than just the pure topologicalSort function tested in
// topsort_test.go: many fresh composers, each seeded with the modules in a
// deliberately "wrong" registration order, must all produce the identical
// RegisterServices and InitializeFromModules sequence.
func TestComposer_Initialize_DeterministicOrderAcrossRuns(t *testing.T) {
	expectedRegister := []string{"a-register", "b-register", "c-register", "d-register"}
	expectedInit := []string{"a-init", "b-init", "c-init", "d-init"}

	for iteration := 0; iteration < 100; iteration++ {
		var order []string

		c := newTestComposer(context.Background())

		a := newOrderRecordingModule("a", nil, &order)
		b := newOrderRecordingModule("b", []string{"a"}, &order)
		cc := newOrderRecordingModule("c", []string{"a"}, &order)
		d := newOrderRecordingModule("d", []string{"b", "c"}, &order)

		c.MustAddModules(d, cc, b, a)

		require.NoError(t, c.Initialize())
		assert.Equal(t, expectedRegister, filterSuffix(order, "-register"))
		assert.Equal(t, expectedInit, filterSuffix(order, "-init"))
	}
}

// TestComposer_Start_UsesInitializeOrderForStartPass proves the
// StartableModule pass, run from Start(), also observes dependency order -
// closing the same map-order gap for starters that Initialize closes for
// RegisterServices/InitializeFromModules.
func TestComposer_Start_UsesInitializeOrderForStartPass(t *testing.T) {
	for iteration := 0; iteration < 20; iteration++ {
		var order []string

		ctx, cancel := context.WithCancel(context.Background())

		c := newTestComposer(ctx)

		a := newOrderRecordingModule("a", nil, &order)
		b := newOrderRecordingModule("b", []string{"a"}, &order)

		c.MustAddModules(b, a)

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		require.NoError(t, c.Start())
		assert.Equal(t, []string{"a-start", "b-start"}, filterSuffix(order, "-start"))
	}
}

// publishedValue is the value cloudsqlconn-style modules provide to the
// injector from within InitializeFromModules, for a dependent to consume
// from within its own InitializeFromModules in the same pass.
type publishedValue struct {
	value string
}

type publisherModule struct {
	name string
}

func (m *publisherModule) Name() string                        { return m.name }
func (m *publisherModule) DependsOn() []string                 { return nil }
func (m *publisherModule) RegisterServices() func(do.Injector) { return nil }

func (m *publisherModule) InitializeFromModules(_ context.Context, injector do.Injector, _ []module.Module) error {
	do.ProvideValue(injector, publishedValue{value: "from-" + m.name})

	return nil
}

type consumerModule struct {
	name      string
	dependsOn []string
	observed  string
}

func (m *consumerModule) Name() string                        { return m.name }
func (m *consumerModule) DependsOn() []string                 { return m.dependsOn }
func (m *consumerModule) RegisterServices() func(do.Injector) { return nil }

func (m *consumerModule) InitializeFromModules(_ context.Context, injector do.Injector, _ []module.Module) error {
	value, err := do.Invoke[publishedValue](injector)
	if err != nil {
		return err
	}

	m.observed = value.value

	return nil
}

// TestComposer_Initialize_DependentObservesDependencyPublishInSamePass
// mirrors the real-world go-foundation defect this fix targets: cloudsqlconn
// publishes a value from its own InitializeFromModules and sqldb resolves it
// from its own InitializeFromModules, in the very same pass. The dependent
// must observe the published value every run, even though it is registered
// before its dependency.
func TestComposer_Initialize_DependentObservesDependencyPublishInSamePass(t *testing.T) {
	for iteration := 0; iteration < 50; iteration++ {
		c := newTestComposer(context.Background())

		pub := &publisherModule{name: "publisher"}
		con := &consumerModule{name: "consumer", dependsOn: []string{"publisher"}}

		c.MustAddModules(con, pub)

		require.NoError(t, c.Initialize())
		assert.Equal(t, "from-publisher", con.observed)
	}
}

func TestComposer_Initialize_DependencyCycle(t *testing.T) {
	c := newTestComposer(context.Background())

	a := newTestModule("a", []string{"b"}, "")
	b := newTestModule("b", []string{"a"}, "")

	c.MustAddModules(a, b)

	err := c.Initialize()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDependencyCycle)
	assert.Contains(t, err.Error(), "a -> b -> a")
}

// TestComposer_Start_RejectsModulesRegisteredAfterInitialize guards the
// regression that caching orderedModules for the StartableModule pass could
// otherwise introduce: a module added after Initialize() already ran would
// silently never be started (it isn't in the cached order and was never
// registered/initialized either), which is worse than the map-order race
// this whole change fixes. Start must fail loudly instead.
func TestComposer_Start_RejectsModulesRegisteredAfterInitialize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := newTestComposer(ctx)

	a := newTestModule("a", nil, "")
	c.MustAddModule(a)

	require.NoError(t, c.Initialize())

	late := newTestModule("late", nil, "")
	c.MustAddModule(late)

	err := c.Start()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrModuleNotFound)
}
