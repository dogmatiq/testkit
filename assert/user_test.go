package assert_test

import (
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
		fn func(*S),
		expectOk bool,
		expectReport ...string,
	) {
		runTest(
			app,
			func(t *testkit.Test) {
				t.Expect(
					testkit.ExecuteCommand(MessageA{}),
					Should("<criteria>", fn),
				)
			},
			nil, // options
			expectOk,
			expectReport,
		)
	}

	Describe("func Should()", func() {
		DescribeTable(
			"assertion reports",
			test,
			Entry(
				"assertion passed",
				func(*S) {},
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <criteria>`,
			),
			Entry(
				"assertion failed",
				func(s *S) {
					s.Fail()
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called at user_test.go:70`,
			),
			Entry(
				"assertion skipped",
				func(s *S) {
					s.SkipNow()
				},
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <criteria> (the user-defined assertion was skipped)`,
				``,
				`  | EXPLANATION`,
				`  |     SkipNow() called at user_test.go:83`,
			),
			Entry(
				"assertion logged a message with Log()",
				func(s *S) {
					s.Log("<message>")
				},
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <criteria>`,
				``,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
			Entry(
				"assertion logged a message with Logf()",
				func(s *S) {
					s.Logf("<format %s>", "value")
				},
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <criteria>`,
				``,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
			Entry(
				"assertion logged a message with Error()",
				func(s *S) {
					s.Error("<message>")
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Error() called at user_test.go:122`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
			Entry(
				"assertion logged a message with Errorf()",
				func(s *S) {
					s.Errorf("<format %s>", "value")
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Errorf() called at user_test.go:138`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
			Entry(
				"assertion logged a message with Fatal()",
				func(s *S) {
					s.Fatal("<message>")
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fatal() called at user_test.go:154`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
			Entry(
				"assertion logged a message with Fatalf()",
				func(s *S) {
					s.Fatalf("<format %s>", "value")
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fatalf() called at user_test.go:170`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
			Entry(
				"assertion failed within helper",
				func(s *S) {
					helper := func() {
						s.Helper()
						s.Fail()
					}

					helper()
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called indirectly by call at user_test.go:191`,
			),
			Entry(
				"assertion failed with fn marked as a helper",
				func(s *S) {
					s.Helper()
					s.Fail()
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called at user_test.go:205`,
			),
			Entry(
				"assertion failed within helper, with fn also marked as a helper",
				func(s *S) {
					s.Helper()

					helper := func() {
						s.Helper()
						s.Fail()
					}

					helper()
				},
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ <criteria> (the user-defined assertion failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called indirectly by call at user_test.go:225`,
			),
		)
	})

	Describe("type S", func() {
		var run func(func(*S)) *testingmock.T

		BeforeEach(func() {
			run = func(fn func(*S)) *testingmock.T {
				t := &testingmock.T{
					FailSilently: true,
				}

				testkit.
					New(app).
					Begin(t).
					Expect(
						testkit.ExecuteCommand(MessageA{}),
						Should("<criteria>", fn),
					)

				return t
			}
		})

		Describe("func Cleanup()", func() {
			It("registers a function to be executed when the test ends", func() {
				var order []int

				run(func(s *S) {
					s.Cleanup(func() {
						order = append(order, 1)
					})

					s.Cleanup(func() {
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
				run(func(s *S) {
					s.Error()
					gomega.Expect(s.Failed()).To(gomega.BeTrue())
				})
			})

			It("does not abort execution", func() {
				completed := false
				run(func(s *S) {
					s.Error()
					completed = true
				})

				gomega.Expect(completed).To(gomega.BeTrue())
			})
		})

		Describe("func Errorf()", func() {
			It("marks the test as failed", func() {
				run(func(s *S) {
					s.Errorf("<format>")
					gomega.Expect(s.Failed()).To(gomega.BeTrue())
				})
			})

			It("does not abort execution", func() {
				completed := false
				run(func(s *S) {
					s.Errorf("<format>")
					completed = true
				})

				gomega.Expect(completed).To(gomega.BeTrue())
			})
		})

		Describe("func Fail()", func() {
			It("marks the test as failed", func() {
				run(func(s *S) {
					s.Fail()
					gomega.Expect(s.Failed()).To(gomega.BeTrue())
				})
			})

			It("does not abort execution", func() {
				completed := false
				run(func(s *S) {
					s.Fail()
					completed = true
				})

				gomega.Expect(completed).To(gomega.BeTrue())
			})
		})

		Describe("func FailNow()", func() {
			It("marks the test as failed", func() {
				run(func(s *S) {
					defer func() {
						gomega.Expect(s.Failed()).To(gomega.BeTrue())
					}()

					s.FailNow()
				})
			})

			It("aborts execution", func() {
				run(func(s *S) {
					s.FailNow()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Fatal()", func() {
			It("marks the test as failed", func() {
				run(func(s *S) {
					defer func() {
						gomega.Expect(s.Failed()).To(gomega.BeTrue())
					}()

					s.Fatal()
				})
			})

			It("aborts execution", func() {
				run(func(s *S) {
					s.Fatal()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Fatalf()", func() {
			It("marks the test as failed", func() {
				run(func(s *S) {
					defer func() {
						gomega.Expect(s.Failed()).To(gomega.BeTrue())
					}()

					s.Fatalf("<format>")
				})
			})

			It("aborts execution", func() {
				run(func(s *S) {
					s.Fatalf("<format>")
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Parallel()", func() {
			It("does not panic", func() {
				run(func(s *S) {
					gomega.Expect(func() {
						s.Parallel()
					}).NotTo(gomega.Panic())
				})
			})
		})

		Describe("func Name()", func() {
			It("returns the criteria string", func() {
				run(func(s *S) {
					gomega.Expect(s.Name()).To(gomega.Equal("<criteria>"))
				})
			})
		})

		Describe("func Skip()", func() {
			It("marks the test as skipped", func() {
				run(func(s *S) {
					defer func() {
						gomega.Expect(s.Skipped()).To(gomega.BeTrue())
					}()

					s.Skip()
				})
			})

			It("prevents a failure from taking effect", func() {
				t := run(func(s *S) {
					s.Fail()
					s.Skip()
				})

				gomega.Expect(t.Failed()).To(gomega.BeFalse())
			})

			It("aborts execution", func() {
				run(func(s *S) {
					s.Skip()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func SkipNow(", func() {
			It("marks the test as skipped", func() {
				run(func(s *S) {
					defer func() {
						gomega.Expect(s.Skipped()).To(gomega.BeTrue())
					}()

					s.SkipNow()
				})
			})

			It("prevents a failure from taking effect", func() {
				t := run(func(s *S) {
					s.Fail()
					s.SkipNow()
				})

				gomega.Expect(t.Failed()).To(gomega.BeFalse())
			})

			It("aborts execution", func() {
				run(func(s *S) {
					s.SkipNow()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Skipf()", func() {
			It("marks the test as skipped", func() {
				run(func(s *S) {
					defer func() {
						gomega.Expect(s.Skipped()).To(gomega.BeTrue())
					}()

					s.Skipf("<format>")
				})
			})

			It("prevents a failure from taking effect", func() {
				t := run(func(s *S) {
					s.Fail()
					s.Skipf("<format>")
				})

				gomega.Expect(t.Failed()).To(gomega.BeFalse())
			})

			It("aborts execution", func() {
				run(func(s *S) {
					s.Skipf("<format>")
					Fail("execution was not aborted")
				})
			})
		})
	})
})
