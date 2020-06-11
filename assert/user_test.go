package assert_test

import (
	"strings"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega"
)

var _ = Context("user assertions", func() {
	var app dogma.Application

	BeforeEach(func() {
		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "<aggregate-key>")
						c.ConsumesCommandType(MessageA{})
						c.ProducesEventType(MessageB{})
					},
					RouteCommandToInstanceFunc: func(dogma.Message) string {
						return "<aggregate-instance>"
					},
				})
			},
		}
	})

	test := func(
		fn func(*T),
		ok bool,
		report ...string,
	) {
		t := &testingmock.T{
			FailSilently: true,
		}

		testkit.
			New(app).
			Begin(t, testkit.Verbose(false)).
			ExecuteCommand(
				MessageA{},
				Should("<criteria>", fn),
			)

		logs := strings.TrimSpace(strings.Join(t.Logs, "\n"))
		lines := strings.Split(logs, "\n")

		gomega.Expect(lines).To(gomega.Equal(report))
		gomega.Expect(t.Failed).To(gomega.Equal(!ok))
	}

	Describe("func Should()", func() {
		DescribeTable(
			"assertion reports",
			test,
			Entry(
				"assertion passed",
				func(*T) {},
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <criteria>`,
			),
			Entry(
				"assertion failed",
				func(t *T) {
					t.Fail()
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called at user_test.go:77`,
			),
			Entry(
				"assertion skipped",
				func(t *T) {
					t.SkipNow()
				},
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <criteria> (the user-defined assertion was skipped)`,
			),
			Entry(
				"assertion logged a message with Log()",
				func(t *T) {
					t.Log("<message>")
					t.Fail() // must fail for log to be shown
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called at user_test.go:101`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
			Entry(
				"assertion logged a message with Logf()",
				func(t *T) {
					t.Logf("<format %s>", "value")
					t.Fail() // must fail for log to be shown
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called at user_test.go:118`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
			Entry(
				"assertion logged a message with Error()",
				func(t *T) {
					t.Error("<message>")
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Error() called at user_test.go:134`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
			Entry(
				"assertion logged a message with Errorf()",
				func(t *T) {
					t.Errorf("<format %s>", "value")
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Errorf() called at user_test.go:150`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
			Entry(
				"assertion logged a message with Fatal()",
				func(t *T) {
					t.Fatal("<message>")
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fatal() called at user_test.go:166`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
			Entry(
				"assertion logged a message with Fatalf()",
				func(t *T) {
					t.Fatalf("<format %s>", "value")
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fatalf() called at user_test.go:182`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
			Entry(
				"assertion failed within helper",
				func(t *T) {
					helper := func() {
						t.Helper()
						t.Fail()
					}

					helper()
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called indirectly by call at user_test.go:203`,
			),
			Entry(
				"assertion failed with fn marked as a helper",
				func(t *T) {
					t.Helper()
					t.Fail()
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called at user_test.go:217`,
			),
			Entry(
				"assertion failed within helper, with fn also marked as a helper",
				func(t *T) {
					t.Helper()

					helper := func() {
						t.Helper()
						t.Fail()
					}

					helper()
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called indirectly by call at user_test.go:237`,
			),
		)
	})

	Describe("type T", func() {
		var run func(func(*T)) *testingmock.T

		BeforeEach(func() {
			run = func(fn func(*T)) *testingmock.T {
				t := &testingmock.T{
					FailSilently: true,
				}

				testkit.
					New(app).
					Begin(t).
					ExecuteCommand(
						MessageA{},
						Should("<criteria>", fn),
					)

				return t
			}
		})

		Describe("func Cleanup()", func() {
			It("registers a function to be executed when the test ends", func() {
				var order []int

				run(func(t *T) {
					t.Cleanup(func() {
						order = append(order, 1)
					})

					t.Cleanup(func() {
						order = append(order, 2)
					})
				})

				gomega.Expect(order).To(gomega.Equal(
					[]int{2, 1},
				))
			})
		})

		Describe("func Error()", func() {
			It("marks the test as failed", func() {
				run(func(t *T) {
					t.Error()
					gomega.Expect(t.Failed()).To(gomega.BeTrue())
				})
			})

			It("does not abort execution", func() {
				completed := false
				run(func(t *T) {
					t.Error()
					completed = true
				})

				gomega.Expect(completed).To(gomega.BeTrue())
			})
		})

		Describe("func Errorf()", func() {
			It("marks the test as failed", func() {
				run(func(t *T) {
					t.Errorf("<format>")
					gomega.Expect(t.Failed()).To(gomega.BeTrue())
				})
			})

			It("does not abort execution", func() {
				completed := false
				run(func(t *T) {
					t.Errorf("<format>")
					completed = true
				})

				gomega.Expect(completed).To(gomega.BeTrue())
			})
		})

		Describe("func Fail()", func() {
			It("marks the test as failed", func() {
				run(func(t *T) {
					t.Fail()
					gomega.Expect(t.Failed()).To(gomega.BeTrue())
				})
			})

			It("does not abort execution", func() {
				completed := false
				run(func(t *T) {
					t.Fail()
					completed = true
				})

				gomega.Expect(completed).To(gomega.BeTrue())
			})
		})

		Describe("func FailNow()", func() {
			It("marks the test as failed", func() {
				run(func(t *T) {
					defer func() {
						gomega.Expect(t.Failed()).To(gomega.BeTrue())
					}()

					t.FailNow()
				})
			})

			It("aborts execution", func() {
				run(func(t *T) {
					t.FailNow()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Fatal()", func() {
			It("marks the test as failed", func() {
				run(func(t *T) {
					defer func() {
						gomega.Expect(t.Failed()).To(gomega.BeTrue())
					}()

					t.Fatal()
				})
			})

			It("aborts execution", func() {
				run(func(t *T) {
					t.Fatal()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Fatalf()", func() {
			It("marks the test as failed", func() {
				run(func(t *T) {
					defer func() {
						gomega.Expect(t.Failed()).To(gomega.BeTrue())
					}()

					t.Fatalf("<format>")
				})
			})

			It("aborts execution", func() {
				run(func(t *T) {
					t.Fatalf("<format>")
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Parallel()", func() {
			It("does not panic", func() {
				run(func(t *T) {
					gomega.Expect(func() {
						t.Parallel()
					}).NotTo(gomega.Panic())
				})
			})
		})

		Describe("func Name()", func() {
			It("returns the criteria string", func() {
				run(func(t *T) {
					gomega.Expect(t.Name()).To(gomega.Equal("<criteria>"))
				})
			})
		})

		Describe("func Skip()", func() {
			It("marks the test as skipped", func() {
				run(func(t *T) {
					defer func() {
						gomega.Expect(t.Skipped()).To(gomega.BeTrue())
					}()

					t.Skip()
				})
			})

			It("prevents a failure from taking effect", func() {
				t := run(func(t *T) {
					t.Fail()
					t.Skip()
				})

				gomega.Expect(t.Failed).To(gomega.BeFalse())
			})

			It("aborts execution", func() {
				run(func(t *T) {
					t.Skip()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func SkipNow(", func() {
			It("marks the test as skipped", func() {
				run(func(t *T) {
					defer func() {
						gomega.Expect(t.Skipped()).To(gomega.BeTrue())
					}()

					t.SkipNow()
				})
			})

			It("prevents a failure from taking effect", func() {
				t := run(func(t *T) {
					t.Fail()
					t.SkipNow()
				})

				gomega.Expect(t.Failed).To(gomega.BeFalse())
			})

			It("aborts execution", func() {
				run(func(t *T) {
					t.SkipNow()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Skipf()", func() {
			It("marks the test as skipped", func() {
				run(func(t *T) {
					defer func() {
						gomega.Expect(t.Skipped()).To(gomega.BeTrue())
					}()

					t.Skipf("<format>")
				})
			})

			It("prevents a failure from taking effect", func() {
				t := run(func(t *T) {
					t.Fail()
					t.Skipf("<format>")
				})

				gomega.Expect(t.Failed).To(gomega.BeFalse())
			})

			It("aborts execution", func() {
				run(func(t *T) {
					t.Skipf("<format>")
					Fail("execution was not aborted")
				})
			})
		})
	})
})
