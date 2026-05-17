package xtesting

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/location"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

// TestingT is the subset of [testing.TB] needed by these helpers.
type TestingT interface {
	Helper()
	Fatal(args ...any)
	Log(args ...any)
	Failed() bool
}

// defaultOptions returns the default cmp.Options applied to every comparison,
// merged with any caller-provided options.
func defaultOptions(options []cmp.Option) []cmp.Option {
	return append(
		[]cmp.Option{
			protocmp.Transform(),
			cmpopts.EquateEmpty(),
			cmpopts.EquateErrors(),
			cmp.Comparer(func(a, b message.Type) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.ApplicationStub) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.AggregateMessageHandlerStub[*stubs.AggregateRootStub]) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.ProcessMessageHandlerStub[*stubs.ProcessRootStub]) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.IntegrationMessageHandlerStub) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.ProjectionMessageHandlerStub) bool { return a == b }),
			cmp.Comparer(func(a, b *config.Aggregate) bool { return a == b }),
			cmp.Comparer(func(a, b *config.Process) bool { return a == b }),
			cmp.Comparer(func(a, b *config.Integration) bool { return a == b }),
			cmp.Comparer(func(a, b *config.Projection) bool { return a == b }),
			cmp.Exporter(
				func(t reflect.Type) bool {
					return t.PkgPath() == "github.com/dogmatiq/enginekit/optional"
				},
			),
		},
		options...,
	)
}

// ExpectContains asserts that slice contains an element equal to want.
func ExpectContains[T any](
	t TestingT,
	failMessage string,
	slice []T,
	want T,
	options ...cmp.Option,
) {
	t.Helper()

	options = defaultOptions(options)

	for _, item := range slice {
		if cmp.Diff(want, item, options...) == "" {
			return
		}
	}

	w := &strings.Builder{}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "===", failMessage, "===")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--- want ---")
	fmt.Fprintln(w, want)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "--- slice ---")
	fmt.Fprintln(w, slice)

	t.Fatal(w.String())
}

// Expect compares two values and fails the test if they are different.
func Expect(
	t TestingT,
	failMessage string,
	got, want any,
	options ...cmp.Option,
) {
	t.Helper()

	options = defaultOptions(options)

	if diff := cmp.Diff(want, got, options...); diff != "" {
		w := &strings.Builder{}

		fmt.Fprintln(w)
		fmt.Fprintln(w, "===", failMessage, "===")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "--- got ---")
		fmt.Fprintln(w, got)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "--- want ---")
		fmt.Fprintln(w, want)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "--- diff ---")
		fmt.Fprintln(w, diff)

		t.Fatal(w.String())
	}
}

// ExpectPanic asserts that a function panics with a specific value.
func ExpectPanic(
	t TestingT,
	want string,
	fn func(),
) {
	t.Helper()

	defer func() {
		t.Helper()

		got := recover()
		if got == nil {
			t.Fatal("expected a panic")
			return
		}

		Expect(
			t,
			"unexpected panic",
			fmt.Sprint(got),
			want,
		)
	}()

	fn()
}

// ExpectPanicMatching asserts that fn panics with a value of type T, then
// calls match to make further assertions about the panic value.
func ExpectPanicMatching[T any](
	t TestingT,
	fn func(),
	match func(T),
) {
	t.Helper()

	defer func() {
		t.Helper()

		r := recover()
		if r == nil {
			t.Fatal("expected a panic")
			return
		}

		v, ok := r.(T)
		if !ok {
			t.Fatal(
				fmt.Sprintf(
					"expected a panic of type %s, got %T",
					reflect.TypeFor[T](),
					r,
				),
			)
			return
		}

		match(v)
	}()

	fn()
}

// ExpectLocation asserts that loc has a non-empty Func, a non-zero Line, and
// that its File ends with the given suffix.
func ExpectLocation(
	t TestingT,
	loc location.Location,
	fileSuffix string,
) {
	t.Helper()

	if loc.Func == "" {
		t.Fatal("expected func to be set in location")
	}

	if !strings.HasSuffix(loc.File, fileSuffix) {
		t.Fatal(fmt.Sprintf("unexpected file in location: got %s, want suffix %s", loc.File, fileSuffix))
	}

	if loc.Line == 0 {
		t.Fatal("expected line to be set in location")
	}
}
