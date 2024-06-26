package testkit

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/iago/must"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/fact"
)

// Test contains the state of a single test.
type Test struct {
	ctx              context.Context
	testingT         TestingT
	app              configkit.RichApplication
	virtualClock     time.Time
	engine           *engine.Engine
	executor         CommandExecutor
	predicateOptions PredicateOptions
	operationOptions []engine.OperationOption
}

// Begin starts a new test.
func Begin(
	t TestingT,
	app dogma.Application,
	options ...TestOption,
) *Test {
	return BeginContext(
		context.Background(),
		t,
		app,
		options...,
	)
}

// BeginContext starts a new test within a context.
func BeginContext(
	ctx context.Context,
	t TestingT,
	app dogma.Application,
	options ...TestOption,
) *Test {
	cfg := configkit.FromApplication(app)

	test := &Test{
		ctx:          ctx,
		testingT:     t,
		app:          cfg,
		virtualClock: time.Now(),
		engine: engine.MustNew(
			cfg,
			engine.EnableProjectionCompactionDuringHandling(true),
		),
		operationOptions: []engine.OperationOption{
			engine.EnableProjections(false),
			engine.EnableIntegrations(false),
			engine.WithObserver(
				fact.NewLogger(func(s string) {
					log(t, s)
				}),
			),
		},
	}

	for _, opt := range options {
		opt.applyTestOption(test)
	}

	return test
}

// Prepare performs a group of actions without making any expectations. It is
// used to place the application into a particular state.
func (t *Test) Prepare(actions ...Action) *Test {
	t.testingT.Helper()

	for _, act := range actions {
		logf(t.testingT, "--- %s ---", act.Caption())
		if err := t.doAction(act); err != nil {
			t.testingT.Fatal(err)
		}
	}

	return t
}

// Expect ensures that a single action results in some expected behavior.
func (t *Test) Expect(act Action, e Expectation) *Test {
	t.testingT.Helper()

	s := PredicateScope{
		App:     t.app,
		Options: t.predicateOptions,
	}

	act.ConfigurePredicate(&s.Options)

	logf(t.testingT, "--- expect %s %s ---", act.Caption(), e.Caption())

	p, err := e.Predicate(s)
	if err != nil {
		t.testingT.Fatal(err)
		return t // required when using a mock testingT that does not panic
	}

	// Using a defer inside a closure satisfies the requirements of the
	// Expectation and Predicate interfaces which state that p.Done() must
	// be called exactly once, and that it must be called before calling
	// p.Report().
	if err := func() error {
		defer p.Done()
		return t.doAction(act, engine.WithObserver(p))
	}(); err != nil {
		t.testingT.Fatal(err)
		return t // required when using a mock testingT that does not panic
	}

	treeOk := p.Ok()
	rep := p.Report(treeOk)

	buf := &strings.Builder{}
	fmt.Fprint(buf, "--- TEST REPORT ---\n\n")
	must.WriteTo(buf, rep)
	t.testingT.Log(buf.String())

	if !treeOk {
		t.testingT.FailNow()
	}

	return t
}

// CommandExecutor returns a dogma.CommandExecutor which can be used to execute
// commands within the context of this test.
//
// The executor can be obtained at any time, but it can only be used within
// specific test actions.
//
// Call() is the only built-in action that supports the command executor. It may
// be supported by other user-defined actions.
func (t *Test) CommandExecutor() dogma.CommandExecutor {
	return &t.executor
}

// EnableHandlers enables a set of handlers by name.
//
// It panics if any of the handler names are not recognized.
//
// By default all integration and projection handlers are disabled.
func (t *Test) EnableHandlers(names ...string) *Test {
	return t.enableHandlers(names, true)
}

// DisableHandlers disables a set of handlers by name.
//
// It panics if any of the handler names are not recognized.
//
// By default all integration and projection handlers are disabled.
func (t *Test) DisableHandlers(names ...string) *Test {
	return t.enableHandlers(names, false)
}

// EnableHandlersLike enables any handlers with a name that matches one of
// the given regular expression patterns.
//
// It panics if any of patterns do not match any handlers.
//
// By default all integration and projection handlers are disabled.
func (t *Test) EnableHandlersLike(patterns ...string) *Test {
	return t.enableHandlersLike(patterns, true)
}

// DisableHandlersLike enables any handlers with a name that matches one of
// the given regular expression patterns.
//
// It panics if any of patterns do not match any handlers.
//
// By default all integration and projection handlers are disabled.
func (t *Test) DisableHandlersLike(patterns ...string) *Test {
	return t.enableHandlersLike(patterns, false)
}

func (t *Test) enableHandlers(names []string, enable bool) *Test {
	for _, n := range names {
		if _, ok := t.app.Handlers().ByName(n); !ok {
			panic(fmt.Sprintf(
				"the %q application does not have a handler named %q",
				t.app.Identity().Name,
				n,
			))
		}

		t.operationOptions = append(
			t.operationOptions,
			engine.EnableHandler(n, enable),
		)
	}

	return t
}

func (t *Test) enableHandlersLike(patterns []string, enable bool) *Test {
	names := map[string]struct{}{}

	for _, p := range patterns {
		re := regexp.MustCompile(p)
		matched := false

		for ident := range t.app.Handlers() {
			if re.MatchString(ident.Name) {
				names[ident.Name] = struct{}{}
				matched = true
			}
		}

		if !matched {
			panic(fmt.Sprintf(
				"the %q application does not have any handlers with names that match the regular expression: %s",
				t.app.Identity().Name,
				p,
			))
		}
	}

	for n := range names {
		t.operationOptions = append(
			t.operationOptions,
			engine.EnableHandler(n, enable),
		)
	}

	return t
}

// doAction calls act.Do() with a scope appropriate for this test.
func (t *Test) doAction(act Action, options ...engine.OperationOption) error {
	opts := []engine.OperationOption{
		engine.WithCurrentTime(t.virtualClock),
	}
	opts = append(opts, t.operationOptions...)
	opts = append(opts, options...)

	return act.Do(
		t.ctx,
		ActionScope{
			App:              t.app,
			VirtualClock:     &t.virtualClock,
			Engine:           t.engine,
			Executor:         &t.executor,
			OperationOptions: opts,
		},
	)
}
