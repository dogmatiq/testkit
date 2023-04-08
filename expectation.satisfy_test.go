package testkit_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("func ToSatisfy()", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
		test     *Test
	)

	BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "04061ede-3f5d-429c-9c14-b140f1cb80c0")
			},
		}

		test = testkit.Begin(testingT, app)
	})

	DescribeTable(
		"expectation behavior",
		func(
			x func(*SatisfyT),
			ok bool,
			rm reportMatcher,
		) {
			test.Expect(
				noop,
				ToSatisfy("<description>", x),
			)

			rm(testingT)
			Expect(testingT.Failed()).To(Equal(!ok))
		},
		Entry(
			"it passes when the expectation does nothing",
			func(*SatisfyT) {},
			expectPass,
			expectReport(
				`✓ <description>`,
			),
		),
		Entry(
			"it fails if Fail() is called",
			func(t *SatisfyT) {
				t.Fail()
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fail() called at expectation.satisfy_test.go:61`,
			),
		),
		Entry(
			"it passes if the expectation is skipped",
			func(t *SatisfyT) {
				t.SkipNow()
			},
			expectPass,
			expectReport(
				`✓ <description> (the expectation was skipped)`,
				``,
				`  | EXPLANATION`,
				`  |     SkipNow() called at expectation.satisfy_test.go:74`,
			),
		),
		Entry(
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
		),
		Entry(
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
		),
		Entry(
			"it fails if Error() is called",
			func(t *SatisfyT) {
				t.Error("<message>")
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Error() called at expectation.satisfy_test.go:113`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
		),
		Entry(
			"fails if Errorf() is called",
			func(t *SatisfyT) {
				t.Errorf("<format %s>", "value")
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Errorf() called at expectation.satisfy_test.go:129`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
		),
		Entry(
			"fails if Fatal() is called",
			func(t *SatisfyT) {
				t.Fatal("<message>")
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fatal() called at expectation.satisfy_test.go:145`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <message>`,
			),
		),
		Entry(
			"fails if Fatalf() is called",
			func(t *SatisfyT) {
				t.Fatalf("<format %s>", "value")
			},
			expectFail,
			expectReport(
				`✗ <description> (the expectation failed)`,
				``,
				`  | EXPLANATION`,
				`  |     Fatalf() called at expectation.satisfy_test.go:161`,
				`  | `,
				`  | LOG MESSAGES`,
				`  |     <format value>`,
			),
		),
		Entry(
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
				`  |     Fail() called indirectly by call at expectation.satisfy_test.go:182`,
			),
		),
		Entry(
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
				`  |     Fail() called at expectation.satisfy_test.go:196`,
			),
		),
		Entry(
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
				`  |     Fail() called indirectly by call at expectation.satisfy_test.go:216`,
			),
		),
	)

	It("produces the expected caption", func() {
		test.Expect(
			noop,
			ToSatisfy(
				"<description>",
				func(*SatisfyT) {},
			),
		)

		Expect(testingT.Logs).To(ContainElement(
			"--- expect [no-op] to <description> ---",
		))
	})

	It("panics if the description is empty", func() {
		Expect(func() {
			ToSatisfy("", func(*SatisfyT) {})
		}).To(PanicWith(`ToSatisfy("", <func>): description must not be empty`))
	})

	It("panics if the function is nil", func() {
		Expect(func() {
			ToSatisfy("<description>", nil)
		}).To(PanicWith(`ToSatisfy("<description>", <nil>): function must not be nil`))
	})

	Describe("type SatisfyT", func() {
		run := func(x func(*SatisfyT)) {
			test.Expect(
				noop,
				ToSatisfy(
					"<description>",
					x,
				),
			)
		}

		Describe("func Cleanup()", func() {
			It("registers a function to be executed when the test ends", func() {
				var order []int

				run(func(t *SatisfyT) {
					t.Cleanup(func() {
						order = append(order, 1)
					})

					t.Cleanup(func() {
						order = append(order, 2)
					})
				})

				Expect(order).To(Equal(
					[]int{2, 1},
				))
			})
		})

		Describe("func Error()", func() {
			It("marks the test as failed", func() {
				run(func(t *SatisfyT) {
					t.Error()
					Expect(t.Failed()).To(BeTrue())
				})
			})

			It("does not abort execution", func() {
				completed := false
				run(func(t *SatisfyT) {
					t.Error()
					completed = true
				})

				Expect(completed).To(BeTrue())
			})
		})

		Describe("func Errorf()", func() {
			It("marks the test as failed", func() {
				run(func(t *SatisfyT) {
					t.Errorf("<format>")
					Expect(t.Failed()).To(BeTrue())
				})
			})

			It("does not abort execution", func() {
				completed := false
				run(func(t *SatisfyT) {
					t.Errorf("<format>")
					completed = true
				})

				Expect(completed).To(BeTrue())
			})
		})

		Describe("func Fail()", func() {
			It("marks the test as failed", func() {
				run(func(t *SatisfyT) {
					t.Fail()
					Expect(t.Failed()).To(BeTrue())
				})
			})

			It("does not abort execution", func() {
				completed := false
				run(func(t *SatisfyT) {
					t.Fail()
					completed = true
				})

				Expect(completed).To(BeTrue())
			})
		})

		Describe("func FailNow()", func() {
			It("marks the test as failed", func() {
				run(func(t *SatisfyT) {
					defer func() {
						Expect(t.Failed()).To(BeTrue())
					}()

					t.FailNow()
				})
			})

			It("aborts execution", func() {
				run(func(t *SatisfyT) {
					t.FailNow()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Fatal()", func() {
			It("marks the test as failed", func() {
				run(func(t *SatisfyT) {
					defer func() {
						Expect(t.Failed()).To(BeTrue())
					}()

					t.Fatal()
				})
			})

			It("aborts execution", func() {
				run(func(t *SatisfyT) {
					t.Fatal()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Fatalf()", func() {
			It("marks the test as failed", func() {
				run(func(t *SatisfyT) {
					defer func() {
						Expect(t.Failed()).To(BeTrue())
					}()

					t.Fatalf("<format>")
				})
			})

			It("aborts execution", func() {
				run(func(t *SatisfyT) {
					t.Fatalf("<format>")
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Parallel()", func() {
			It("does not panic", func() {
				run(func(t *SatisfyT) {
					Expect(func() {
						t.Parallel()
					}).NotTo(Panic())
				})
			})
		})

		Describe("func Name()", func() {
			It("returns the description", func() {
				run(func(t *SatisfyT) {
					Expect(t.Name()).To(Equal("<description>"))
				})
			})
		})

		Describe("func Skip()", func() {
			It("marks the test as skipped", func() {
				run(func(t *SatisfyT) {
					defer func() {
						Expect(t.Skipped()).To(BeTrue())
					}()

					t.Skip()
				})
			})

			It("prevents a failure from taking effect", func() {
				run(func(t *SatisfyT) {
					t.Fail()
					t.Skip()
				})

				Expect(testingT.Failed()).To(BeFalse())
			})

			It("aborts execution", func() {
				run(func(t *SatisfyT) {
					t.Skip()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func SkipNow(", func() {
			It("marks the test as skipped", func() {
				run(func(t *SatisfyT) {
					defer func() {
						Expect(t.Skipped()).To(BeTrue())
					}()

					t.SkipNow()
				})
			})

			It("prevents a failure from taking effect", func() {
				run(func(t *SatisfyT) {
					t.Fail()
					t.SkipNow()
				})

				Expect(testingT.Failed()).To(BeFalse())
			})

			It("aborts execution", func() {
				run(func(t *SatisfyT) {
					t.SkipNow()
					Fail("execution was not aborted")
				})
			})
		})

		Describe("func Skipf()", func() {
			It("marks the test as skipped", func() {
				run(func(t *SatisfyT) {
					defer func() {
						Expect(t.Skipped()).To(BeTrue())
					}()

					t.Skipf("<format>")
				})
			})

			It("prevents a failure from taking effect", func() {
				run(func(t *SatisfyT) {
					t.Fail()
					t.Skipf("<format>")
				})

				Expect(testingT.Failed()).To(BeFalse())
			})

			It("aborts execution", func() {
				run(func(t *SatisfyT) {
					t.Skipf("<format>")
					Fail("execution was not aborted")
				})
			})
		})
	})
})
