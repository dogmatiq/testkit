package compare_test

import (
	"testing"

	. "github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
)

var _ Comparator = DefaultComparator{}

func TestDefaultComparator_MessageIsEqual(t *testing.T) {
	c := DefaultComparator{}

	if c.MessageIsEqual(
		fixtures.MessageA{Value: "<foo>"},
		fixtures.MessageB{Value: "<foo>"},
	) {
		t.Fatal("messages of different types compare as equal")
	}

	if c.MessageIsEqual(
		fixtures.MessageA{Value: "<foo>"},
		fixtures.MessageA{Value: "<bar>"},
	) {
		t.Fatal("messages of the same type with different content compare as equal")
	}

	if !c.MessageIsEqual(
		fixtures.MessageA{Value: "<foo>"},
		fixtures.MessageA{Value: "<foo>"},
	) {
		t.Fatal("identical messages compare as inequal")
	}
}

func TestDefaultComparator_AggregateRootIsEqual(t *testing.T) {
	a := &fixtures.AggregateRoot{
		Value: "<foo>",
	}

	b := &fixtures.AggregateRoot{
		Value: "<bar>",
	}

	c := DefaultComparator{}

	if c.AggregateRootIsEqual(a, b) {
		t.Fatal("different aggregate roots compare as equal")
	}

	if !c.AggregateRootIsEqual(a, a) {
		t.Fatal("identical aggregate roots compare as inequal")
	}
}

func TestDefaultComparator_ProcessRootIsEqual(t *testing.T) {
	a := &fixtures.ProcessRoot{
		Value: "<foo>",
	}

	b := &fixtures.ProcessRoot{
		Value: "<bar>",
	}

	c := DefaultComparator{}

	if c.ProcessRootIsEqual(a, b) {
		t.Fatal("different Process roots compare as equal")
	}

	if !c.ProcessRootIsEqual(a, a) {
		t.Fatal("identical Process roots compare as inequal")
	}
}
