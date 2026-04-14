package fact_test

import (
	"testing"
	"time"

	. "github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/test"
)

func TestObserverGroup(t *testing.T) {
	t.Run("func Notify()", func(t *testing.T) {
		t.Run("it notifies each of the observers in the group", func(t *testing.T) {
			f := TickCycleBegun{}

			n := 0
			g := ObserverGroup{
				ObserverFunc(func(of Fact) {
					n++
					test.Expect(t, "unexpected fact", of, f)
				}),
				ObserverFunc(func(of Fact) {
					n++
					test.Expect(t, "unexpected fact", of, f)
				}),
			}

			g.Notify(f)

			test.Expect(t, "unexpected notification count", n, 2)
		})
	})
}

func TestBuffer(t *testing.T) {
	t.Run("func Notify()", func(t *testing.T) {
		t.Run("it appends the fact to the buffer", func(t *testing.T) {
			f1 := TickCycleBegun{
				EngineTime: time.Now(),
			}
			f2 := TickCycleBegun{
				EngineTime: time.Now().Add(1 * time.Second),
			}
			b := &Buffer{}

			b.Notify(f1)
			b.Notify(f2)

			test.Expect(
				t,
				"unexpected buffered facts",
				b.Facts(),
				[]Fact{f1, f2},
			)
		})
	})
}

func TestIgnore(t *testing.T) {
	t.Run("func Notify()", func(t *testing.T) {
		t.Run("it does nothing", func(t *testing.T) {
			Ignore.Notify(TickCycleBegun{})
		})
	})
}
