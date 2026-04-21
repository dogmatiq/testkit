package testkit_test

import (
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestToSatisfy(t *testing.T) {
	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "04061ede-3f5d-429c-9c14-b140f1cb80c0")
		},
	}

	cases := []struct {
		Name   string
		Func   func(*SatisfyT)
		Passes bool
		Report reportMatcher
	}{
		{
			"it passes when the expectation does nothing",
			func(*SatisfyT) {},
			expectPass,
			expectReport(
				`✓ <description>`,
			),
		},
		{
			"it fails if Fail() is called",
			func(t *SatisfyT) {
				t.Fail()
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called at expectation.satisfy_test.go:37`,
			),
		},
		{
			"it passes if the expectation is skipped",
			func(t *SatisfyT) {
				t.SkipNow()
			},
			expectPass,
			expectReport(
				`✓ <description> (the expectation was skipped)`,
				``,
				`  | EXPLANATION`,
				`  |     SkipNow() called at expectation.satisfy_test.go:50`,
			),
		},
		{
			"it includes Log() messages in the test report",
			func(t *SatisfyT) {
				t.Log("<message>")
			},
			expectPass,
			expectReport(
				`✓ <description>`,
				``,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
		},
		{
			"it includes Logf() messages in the test report",
			func(t *SatisfyT) {
				t.Logf("<format %s>", "value")
			},
			expectPass,
			expectReport(
				`✓ <description>`,
				``,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
		},
		{
			"it fails if Error() is called",
			func(t *SatisfyT) {
				t.Error("<message>")
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Error() called at expectation.satisfy_test.go:89`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
		},
		{
			"fails if Errorf() is called",
			func(t *SatisfyT) {
				t.Errorf("<format %s>", "value")
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Errorf() called at expectation.satisfy_test.go:105`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
		},
		{
			"fails if Fatal() is called",
			func(t *SatisfyT) {
				t.Fatal("<message>")
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fatal() called at expectation.satisfy_test.go:121`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
		},
		{
			"fails if Fatalf() is called",
			func(t *SatisfyT) {
				t.Fatalf("<format %s>", "value")
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fatalf() called at expectation.satisfy_test.go:137`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
		},
		{
			"fails if Fail() is called within a helper function",
			func(t *SatisfyT) {
				helper := func() {
					t.Helper()
					t.Fail()
				}

				helper()
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called indirectly by call at expectation.satisfy_test.go:158`,
			),
		},
		{
			"fails if Fail() is called when the expectation function itself is a helper function",
			func(t *SatisfyT) {
				t.Helper()
				t.Fail()
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called at expectation.satisfy_test.go:172`,
			),
		},
		{
			"fails if Fail() is called within a helper function when the expectation function itself is also a helper function",
			func(t *SatisfyT) {
				t.Helper()

				helper := func() {
					t.Helper()
					t.Fail()
				}

				helper()
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called indirectly by call at expectation.satisfy_test.go:192`,
			),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mt := &testingmock.T{FailSilently: true}
			test := Begin(mt, app)
			test.Expect(noop, ToSatisfy("<description>", c.Func))

			if mt.Failed() != !c.Passes {
				t.Fatalf("testingT.Failed() = %v, want %v", mt.Failed(), !c.Passes)
			}

			preReportCount := len(mt.Logs)
			c.Report(mt)
			if len(mt.Logs) > preReportCount {
				t.Fatalf("report content mismatch:\n%v", mt.Logs[preReportCount:])
			}
		})
	}

	t.Run("does not include an explanation when negated and a sibling expectation passes", func(t *testing.T) {
		mt := &testingmock.T{FailSilently: true}
		test := Begin(mt, app)
		test.Expect(
			noop,
			NoneOf(
				ToSatisfy("<always pass>", func(*SatisfyT) {}),
				ToSatisfy("<always fail>", func(t *SatisfyT) { t.Fail() }),
			),
		)

		rm := expectReport(
			`✗ none of (1 of the expectations passed unexpectedly)`,
			`    ✓ <always pass>`,
			`    ✗ <always fail> (the expectation failed)`,
		)
		rm(mt)

		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
	})

	t.Run("produces the expected caption", func(t *testing.T) {
		mt := &testingmock.T{FailSilently: true}
		test := Begin(mt, app)
		test.Expect(
			noop,
			ToSatisfy(
				"<description>",
				func(*SatisfyT) {},
			),
		)

		var found bool
		for _, log := range mt.Logs {
			if log == "--- expect [no-op] to <description> ---" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected log message not found, got: %v", mt.Logs)
		}
	})

	t.Run("panics if the description is empty", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			`ToSatisfy("", <func>): description must not be empty`,
			func() {
				ToSatisfy("", func(*SatisfyT) {})
			},
		)
	})

	t.Run("panics if the function is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			`ToSatisfy("<description>", <nil>): function must not be nil`,
			func() {
				ToSatisfy("<description>", nil)
			},
		)
	})

	t.Run("type SatisfyT", func(t *testing.T) {
		type env struct {
			mt  *testingmock.T
			run func(func(*SatisfyT))
		}

		newEnv := func() *env {
			mt := &testingmock.T{FailSilently: true}
			test := Begin(mt, app)
			return &env{
				mt: mt,
				run: func(x func(*SatisfyT)) {
					test.Expect(noop, ToSatisfy("<description>", x))
				},
			}
		}

		t.Run("func Cleanup()", func(t *testing.T) {
			t.Run("registers a function to be executed when the test ends", func(t *testing.T) {
				e := newEnv()
				var order []int

				e.run(func(st *SatisfyT) {
					st.Cleanup(func() {
						order = append(order, 1)
					})

					st.Cleanup(func() {
						order = append(order, 2)
					})
				})

				xtesting.Expect(t, "cleanup order", order, []int{2, 1})
			})
		})

		t.Run("func Error()", func(t *testing.T) {
			t.Run("marks the test as failed", func(t *testing.T) {
				e := newEnv()
				var failed bool

				e.run(func(st *SatisfyT) {
					st.Error()
					failed = st.Failed()
				})

				if !failed {
					t.Fatal("expected st.Failed() to be true after Error()")
				}
			})

			t.Run("does not abort execution", func(t *testing.T) {
				e := newEnv()
				completed := false

				e.run(func(st *SatisfyT) {
					st.Error()
					completed = true
				})

				if !completed {
					t.Fatal("expected execution to continue after Error()")
				}
			})
		})

		t.Run("func Errorf()", func(t *testing.T) {
			t.Run("marks the test as failed", func(t *testing.T) {
				e := newEnv()
				var failed bool

				e.run(func(st *SatisfyT) {
					st.Errorf("<format>")
					failed = st.Failed()
				})

				if !failed {
					t.Fatal("expected st.Failed() to be true after Errorf()")
				}
			})

			t.Run("does not abort execution", func(t *testing.T) {
				e := newEnv()
				completed := false

				e.run(func(st *SatisfyT) {
					st.Errorf("<format>")
					completed = true
				})

				if !completed {
					t.Fatal("expected execution to continue after Errorf()")
				}
			})
		})

		t.Run("func Fail()", func(t *testing.T) {
			t.Run("marks the test as failed", func(t *testing.T) {
				e := newEnv()
				var failed bool

				e.run(func(st *SatisfyT) {
					st.Fail()
					failed = st.Failed()
				})

				if !failed {
					t.Fatal("expected st.Failed() to be true after Fail()")
				}
			})

			t.Run("does not abort execution", func(t *testing.T) {
				e := newEnv()
				completed := false

				e.run(func(st *SatisfyT) {
					st.Fail()
					completed = true
				})

				if !completed {
					t.Fatal("expected execution to continue after Fail()")
				}
			})
		})

		t.Run("func FailNow()", func(t *testing.T) {
			t.Run("marks the test as failed", func(t *testing.T) {
				e := newEnv()
				var failed bool

				e.run(func(st *SatisfyT) {
					defer func() { failed = st.Failed() }()
					st.FailNow()
				})

				if !failed {
					t.Fatal("expected st.Failed() to be true after FailNow()")
				}
			})

			t.Run("aborts execution", func(t *testing.T) {
				e := newEnv()
				aborted := true

				e.run(func(st *SatisfyT) {
					st.FailNow()
					aborted = false
				})

				if !aborted {
					t.Fatal("execution was not aborted")
				}
			})
		})

		t.Run("func Fatal()", func(t *testing.T) {
			t.Run("marks the test as failed", func(t *testing.T) {
				e := newEnv()
				var failed bool

				e.run(func(st *SatisfyT) {
					defer func() { failed = st.Failed() }()
					st.Fatal()
				})

				if !failed {
					t.Fatal("expected st.Failed() to be true after Fatal()")
				}
			})

			t.Run("aborts execution", func(t *testing.T) {
				e := newEnv()
				aborted := true

				e.run(func(st *SatisfyT) {
					st.Fatal()
					aborted = false
				})

				if !aborted {
					t.Fatal("execution was not aborted")
				}
			})
		})

		t.Run("func Fatalf()", func(t *testing.T) {
			t.Run("marks the test as failed", func(t *testing.T) {
				e := newEnv()
				var failed bool

				e.run(func(st *SatisfyT) {
					defer func() { failed = st.Failed() }()
					st.Fatalf("<format>")
				})

				if !failed {
					t.Fatal("expected st.Failed() to be true after Fatalf()")
				}
			})

			t.Run("aborts execution", func(t *testing.T) {
				e := newEnv()
				aborted := true

				e.run(func(st *SatisfyT) {
					st.Fatalf("<format>")
					aborted = false
				})

				if !aborted {
					t.Fatal("execution was not aborted")
				}
			})
		})

		t.Run("func Parallel()", func(t *testing.T) {
			t.Run("does not panic", func(t *testing.T) {
				e := newEnv()

				e.run(func(st *SatisfyT) {
					st.Parallel()
				})
			})
		})

		t.Run("func Name()", func(t *testing.T) {
			t.Run("returns the description", func(t *testing.T) {
				e := newEnv()
				var name string

				e.run(func(st *SatisfyT) {
					name = st.Name()
				})

				if name != "<description>" {
					t.Fatalf("expected name %q, got %q", "<description>", name)
				}
			})
		})

		t.Run("func Skip()", func(t *testing.T) {
			t.Run("marks the test as skipped", func(t *testing.T) {
				e := newEnv()
				var skipped bool

				e.run(func(st *SatisfyT) {
					defer func() { skipped = st.Skipped() }()
					st.Skip()
				})

				if !skipped {
					t.Fatal("expected st.Skipped() to be true after Skip()")
				}
			})

			t.Run("prevents a failure from taking effect", func(t *testing.T) {
				e := newEnv()

				e.run(func(st *SatisfyT) {
					st.Fail()
					st.Skip()
				})

				if e.mt.Failed() {
					t.Fatal("expected outer test to pass because Skip() cancels Fail()")
				}
			})

			t.Run("aborts execution", func(t *testing.T) {
				e := newEnv()
				aborted := true

				e.run(func(st *SatisfyT) {
					st.Skip()
					aborted = false
				})

				if !aborted {
					t.Fatal("execution was not aborted")
				}
			})
		})

		t.Run("func SkipNow()", func(t *testing.T) {
			t.Run("marks the test as skipped", func(t *testing.T) {
				e := newEnv()
				var skipped bool

				e.run(func(st *SatisfyT) {
					defer func() { skipped = st.Skipped() }()
					st.SkipNow()
				})

				if !skipped {
					t.Fatal("expected st.Skipped() to be true after SkipNow()")
				}
			})

			t.Run("prevents a failure from taking effect", func(t *testing.T) {
				e := newEnv()

				e.run(func(st *SatisfyT) {
					st.Fail()
					st.SkipNow()
				})

				if e.mt.Failed() {
					t.Fatal("expected outer test to pass because SkipNow() cancels Fail()")
				}
			})

			t.Run("aborts execution", func(t *testing.T) {
				e := newEnv()
				aborted := true

				e.run(func(st *SatisfyT) {
					st.SkipNow()
					aborted = false
				})

				if !aborted {
					t.Fatal("execution was not aborted")
				}
			})
		})

		t.Run("func Skipf()", func(t *testing.T) {
			t.Run("marks the test as skipped", func(t *testing.T) {
				e := newEnv()
				var skipped bool

				e.run(func(st *SatisfyT) {
					defer func() { skipped = st.Skipped() }()
					st.Skipf("<format>")
				})

				if !skipped {
					t.Fatal("expected st.Skipped() to be true after Skipf()")
				}
			})

			t.Run("prevents a failure from taking effect", func(t *testing.T) {
				e := newEnv()

				e.run(func(st *SatisfyT) {
					st.Fail()
					st.Skipf("<format>")
				})

				if e.mt.Failed() {
					t.Fatal("expected outer test to pass because Skipf() cancels Fail()")
				}
			})

			t.Run("aborts execution", func(t *testing.T) {
				e := newEnv()
				aborted := true

				e.run(func(st *SatisfyT) {
					st.Skipf("<format>")
					aborted = false
				})

				if !aborted {
					t.Fatal("execution was not aborted")
				}
			})
		})
	})
}
