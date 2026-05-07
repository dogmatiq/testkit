package compare_test

import (
	"testing"

	"github.com/dogmatiq/testkit/internal/compare"
	"github.com/dogmatiq/testkit/internal/fixtures"
)

func TestEqual(t *testing.T) {
	t.Run("plain struct", func(t *testing.T) {
		type S struct{ X int }

		t.Run("equal", func(t *testing.T) {
			if !compare.Equal(S{1}, S{1}) {
				t.Fatal("expected values to be equal")
			}
		})

		t.Run("not equal", func(t *testing.T) {
			if compare.Equal(S{1}, S{2}) {
				t.Fatal("expected values to be different")
			}
		})
	})

	t.Run("protocol buffers", func(t *testing.T) {
		t.Run("equal", func(t *testing.T) {
			a := fixtures.NewProtoMessageBuilder().WithValue("<value>").Build()
			b := fixtures.NewProtoMessageBuilder().WithValue("<value>").Build()

			if !compare.Equal(a, b) {
				t.Fatal("expected values to be equal")
			}
		})

		t.Run("not equal", func(t *testing.T) {
			a := fixtures.NewProtoMessageBuilder().WithValue("<value-a>").Build()
			b := fixtures.NewProtoMessageBuilder().WithValue("<value-b>").Build()

			if compare.Equal(a, b) {
				t.Fatal("expected values to be different")
			}
		})

		t.Run("ignores unexported fields", func(t *testing.T) {
			a := fixtures.NewProtoMessageBuilder().WithValue("<value>").Build()
			b := fixtures.NewProtoMessageBuilder().WithValue("<value>").Build()

			// Force population of the internal sizeCache field.
			_ = a.String()

			if !compare.Equal(a, b) {
				t.Fatal("expected proto.Equal to be used, ignoring unexported fields")
			}
		})
	})
}
