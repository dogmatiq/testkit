package location_test

import (
	"strings"
	"testing"

	. "github.com/dogmatiq/testkit/location"
	"github.com/dogmatiq/testkit/x/xtesting"
)

func TestLocation(t *testing.T) {
	t.Run("func OfFunc()", func(t *testing.T) {
		t.Run("it returns the expected location", func(t *testing.T) {
			loc := OfFunc(doNothing)

			if got, want := loc.Func, "github.com/dogmatiq/testkit/location_test.doNothing"; got != want {
				t.Fatalf("unexpected function name: got %q, want %q", got, want)
			}

			if !strings.HasSuffix(loc.File, "/location/linenumber_test.go") {
				t.Fatalf("unexpected file name: got %q, want suffix %q", loc.File, "/location/linenumber_test.go")
			}

			if got, want := loc.Line, 50; got != want {
				t.Fatalf("unexpected line number: got %d, want %d", got, want)
			}
		})

		t.Run("it panics if the value is not a function", func(t *testing.T) {
			xtesting.ExpectPanic(t, "fn must be a function", func() {
				OfFunc("<not a function>")
			})
		})
	})

	t.Run("func OfMethod()", func(t *testing.T) {
		t.Run("it returns the expected location", func(t *testing.T) {
			loc := OfMethod(ofMethodT{}, "Method")

			if got, want := loc.Func, "github.com/dogmatiq/testkit/location_test.ofMethodT.Method"; got != want {
				t.Fatalf("unexpected function name: got %q, want %q", got, want)
			}

			if !strings.HasSuffix(loc.File, "/location/linenumber_test.go") {
				t.Fatalf("unexpected file name: got %q, want suffix %q", loc.File, "/location/linenumber_test.go")
			}

			if got, want := loc.Line, 57; got != want {
				t.Fatalf("unexpected line number: got %d, want %d", got, want)
			}
		})

		t.Run("it panics if the methods does not exist", func(t *testing.T) {
			xtesting.ExpectPanic(t, "method does not exist", func() {
				OfMethod(ofMethodT{}, "DoesNotExist")
			})
		})
	})

	t.Run("func OfCall()", func(t *testing.T) {
		t.Run("it returns the expected location", func(t *testing.T) {
			loc := ofCallLayer2()

			if got, want := loc.Func, "github.com/dogmatiq/testkit/location_test.ofCallLayer2"; got != want {
				t.Fatalf("unexpected function name: got %q, want %q", got, want)
			}

			if !strings.HasSuffix(loc.File, "/location/linenumber_test.go") {
				t.Fatalf("unexpected file name: got %q, want suffix %q", loc.File, "/location/linenumber_test.go")
			}

			if got, want := loc.Line, 53; got != want {
				t.Fatalf("unexpected line number: got %d, want %d", got, want)
			}
		})
	})

	t.Run("func OfPanic()", func(t *testing.T) {
		t.Run("it returns the expected location", func(t *testing.T) {
			defer func() {
				recover()
				loc := OfPanic()

				if got, want := loc.Func, "github.com/dogmatiq/testkit/location_test.doPanic"; got != want {
					t.Fatalf("unexpected function name: got %q, want %q", got, want)
				}

				if !strings.HasSuffix(loc.File, "/location/linenumber_test.go") {
					t.Fatalf("unexpected file name: got %q, want suffix %q", loc.File, "/location/linenumber_test.go")
				}

				if got, want := loc.Line, 51; got != want {
					t.Fatalf("unexpected line number: got %d, want %d", got, want)
				}
			}()

			doPanic()
		})
	})

	t.Run("func String()", func(t *testing.T) {
		cases := []struct {
			Name     string
			Location Location
			Want     string
		}{
			{
				Name:     "empty",
				Location: Location{},
				Want:     "<unknown>",
			},
			{
				Name:     "function name only",
				Location: Location{Func: "<function>"},
				Want:     "<function>(...)",
			},
			{
				Name:     "function name only (global closure)",
				Location: Location{Func: "<function glob..>"},
				Want:     "<function glob..>(...)",
			},
			{
				Name:     "file location only",
				Location: Location{File: "<file>", Line: 123},
				Want:     "<file>:123",
			},
			{
				Name:     "both",
				Location: Location{Func: "<function>", File: "<file>", Line: 123},
				Want:     "<file>:123 [<function>(...)]",
			},
			{
				Name:     "both (global closure)",
				Location: Location{Func: "<function glob..>", File: "<file>", Line: 123},
				Want:     "<file>:123",
			},
		}

		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				if got := c.Location.String(); got != c.Want {
					t.Fatalf("unexpected string representation: got %q, want %q", got, c.Want)
				}
			})
		}
	})
}
