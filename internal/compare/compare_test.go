package compare_test

import (
	"testing"

	. "github.com/dogmatiq/testkit/internal/compare"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ExampleType struct {
	Value    string
	CallFunc func()
}

type nestedFuncType struct {
	Value string
	Inner ExampleType
}

type unexportedFieldType struct {
	value string
}

type unexportedWithFuncType struct {
	value    string
	CallFunc func()
}

func TestEqual(t *testing.T) {
	factory := func() func() { return func() {} }

	// fn1 and fn2 are separate instantiations from the same factory,
	// so they share the same closure definition site.
	fn1 := factory()
	fn2 := factory()

	// differentFn is defined at a different source location.
	differentFn := func() {}

	cases := []struct {
		Name  string
		A, B  any
		Equal bool
	}{
		// Scalar values.
		{"equal strings", "foo", "foo", true},
		{"non-equal strings", "foo", "bar", false},
		{"equal ints", 42, 42, true},
		{"non-equal ints", 42, 43, false},

		// Proto messages.
		{"equal proto messages", wrapperspb.String("foo"), wrapperspb.String("foo"), true},
		{"non-equal proto messages", wrapperspb.String("foo"), wrapperspb.String("bar"), false},

		// Functions compared by definition site.
		{"same definition funcs", fn1, fn2, true},
		{"different definition funcs", fn1, differentFn, false},
		{"nil vs non-nil func", (func())(nil), fn1, false},
		{"both nil funcs", (func())(nil), (func())(nil), true},

		// Pointers.
		{"equal pointers", &ExampleType{Value: "x", CallFunc: fn1}, &ExampleType{Value: "x", CallFunc: fn2}, true},
		{"non-equal pointers", &ExampleType{Value: "x"}, &ExampleType{Value: "y"}, false},
		{"nil and non-nil pointer", (*ExampleType)(nil), &ExampleType{}, false},
		{"both nil pointers", (*ExampleType)(nil), (*ExampleType)(nil), true},

		// Structs with func fields compared by location.
		{"same definition func fields", ExampleType{Value: "x", CallFunc: fn1}, ExampleType{Value: "x", CallFunc: fn2}, true},
		{"different definition func fields", ExampleType{Value: "x", CallFunc: fn1}, ExampleType{Value: "x", CallFunc: differentFn}, false},
		{"non-equal data fields", ExampleType{Value: "x"}, ExampleType{Value: "y"}, false},

		// Structs without func fields (includes unexported fields).
		{"unexported fields equal", unexportedFieldType{value: "x"}, unexportedFieldType{value: "x"}, true},
		{"unexported fields non-equal", unexportedFieldType{value: "x"}, unexportedFieldType{value: "y"}, false},

		// Structs with both unexported fields and func fields.
		{"unexported with func equal", unexportedWithFuncType{value: "x", CallFunc: fn1}, unexportedWithFuncType{value: "x", CallFunc: fn2}, true},
		{"unexported with func non-equal value", unexportedWithFuncType{value: "x", CallFunc: fn1}, unexportedWithFuncType{value: "y", CallFunc: fn2}, false},
		{"unexported with func non-equal func", unexportedWithFuncType{value: "x", CallFunc: fn1}, unexportedWithFuncType{value: "x", CallFunc: differentFn}, false},

		// Structs with nested func fields.
		{"nested same definition func fields", nestedFuncType{Value: "x", Inner: ExampleType{CallFunc: fn1}}, nestedFuncType{Value: "x", Inner: ExampleType{CallFunc: fn2}}, true},
		{"nested different definition func fields", nestedFuncType{Value: "x", Inner: ExampleType{CallFunc: fn1}}, nestedFuncType{Value: "x", Inner: ExampleType{CallFunc: differentFn}}, false},
		{"nested non-equal data fields", nestedFuncType{Value: "x"}, nestedFuncType{Value: "y"}, false},

		// Slices.
		{"equal slices", []ExampleType{{Value: "a", CallFunc: fn1}}, []ExampleType{{Value: "a", CallFunc: fn2}}, true},
		{"non-equal slices", []ExampleType{{Value: "a"}}, []ExampleType{{Value: "b"}}, false},
		{"different length slices", []ExampleType{{Value: "a"}}, []ExampleType{}, false},
		{"nil vs non-nil slice", []ExampleType(nil), []ExampleType{}, false},
		{"both nil slices", []ExampleType(nil), []ExampleType(nil), true},

		// Slices of pointers.
		{"equal slices of pointers", []*ExampleType{{Value: "a", CallFunc: fn1}}, []*ExampleType{{Value: "a", CallFunc: fn2}}, true},
		{"non-equal slices of pointers", []*ExampleType{{Value: "a"}}, []*ExampleType{{Value: "b"}}, false},

		// Maps.
		{"equal maps", map[string]ExampleType{"k": {Value: "v", CallFunc: fn1}}, map[string]ExampleType{"k": {Value: "v", CallFunc: fn2}}, true},
		{"non-equal map values", map[string]ExampleType{"k": {Value: "a"}}, map[string]ExampleType{"k": {Value: "b"}}, false},
		{"different map keys", map[string]ExampleType{"a": {}}, map[string]ExampleType{"b": {}}, false},
		{"different map lengths", map[string]ExampleType{"a": {}}, map[string]ExampleType{}, false},
		{"nil vs non-nil map", map[string]ExampleType(nil), map[string]ExampleType{}, false},
		{"both nil maps", map[string]ExampleType(nil), map[string]ExampleType(nil), true},

		// Arrays.
		{"equal arrays", [2]ExampleType{{Value: "a", CallFunc: fn1}, {Value: "b"}}, [2]ExampleType{{Value: "a", CallFunc: fn2}, {Value: "b"}}, true},
		{"non-equal arrays", [1]ExampleType{{Value: "a"}}, [1]ExampleType{{Value: "b"}}, false},

		// Interfaces (via []any).
		{"equal interface elements", []any{&ExampleType{Value: "x", CallFunc: fn1}}, []any{&ExampleType{Value: "x", CallFunc: fn2}}, true},
		{"non-equal interface elements", []any{&ExampleType{Value: "x"}}, []any{&ExampleType{Value: "y"}}, false},
		{"nil interface element", []any{nil}, []any{nil}, true},
		{"nil vs non-nil interface element", []any{nil}, []any{&ExampleType{}}, false},

		// Nil (untyped).
		{"both nil", nil, nil, true},
		{"nil vs non-nil", nil, "x", false},
		{"non-nil vs nil", "x", nil, false},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := Equal(c.A, c.B)
			if got != c.Equal {
				t.Fatalf("Equal() = %v, want %v", got, c.Equal)
			}
		})
	}
}
