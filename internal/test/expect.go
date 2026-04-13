package test

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/enginekit/message"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

// Expect compares two values and fails the test if they are different.
func Expect(
	t *testing.T,
	failMessage string,
	got, want any,
	options ...cmp.Option,
) {
	t.Helper()

	options = append(
		[]cmp.Option{
			protocmp.Transform(),
			cmpopts.EquateEmpty(),
			cmpopts.EquateErrors(),
			cmp.Comparer(func(a, b message.Type) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.ApplicationStub) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.AggregateMessageHandlerStub) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.ProcessMessageHandlerStub) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.IntegrationMessageHandlerStub) bool { return a == b }),
			cmp.Comparer(func(a, b *stubs.ProjectionMessageHandlerStub) bool { return a == b }),
			cmp.Exporter(
				func(t reflect.Type) bool {
					return t.PkgPath() == "github.com/dogmatiq/enginekit/optional"
				},
			),
		},
		options...,
	)

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
	t *testing.T,
	want string,
	fn func(),
) {
	t.Helper()

	defer func() {
		t.Helper()

		got := recover()

		if got != nil {
			defer func() {
				t.Helper()

				if t.Failed() {
					m := "\n=== stack trace ===\n\n" + string(debug.Stack())
					t.Log(m)
				}
			}()
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
