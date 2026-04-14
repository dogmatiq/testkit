package envelope_test

import (
	"testing"

	. "github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestMessageIDGenerator(t *testing.T) {
	t.Run("func Next()", func(t *testing.T) {
		t.Run("it returns the next ID in the sequence", func(t *testing.T) {
			generator := &MessageIDGenerator{}

			xtesting.Expect(t, "unexpected message ID", generator.Next(), "1")
			xtesting.Expect(t, "unexpected message ID", generator.Next(), "2")
			xtesting.Expect(t, "unexpected message ID", generator.Next(), "3")
		})
	})

	t.Run("func Reset()", func(t *testing.T) {
		t.Run("it returns the sequence to 1", func(t *testing.T) {
			generator := &MessageIDGenerator{}

			generator.Next()
			generator.Next()
			generator.Next()
			generator.Reset()

			xtesting.Expect(t, "unexpected message ID", generator.Next(), "1")
		})
	})
}
