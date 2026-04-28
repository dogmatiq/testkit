package xreflect_test

import (
	"testing"

	. "github.com/dogmatiq/testkit/internal/x/xreflect"
)

func TestIsNil(t *testing.T) {
	cases := []struct {
		Name  string
		Value any
		Want  bool
	}{
		{
			Name:  "untyped nil",
			Value: nil,
			Want:  true,
		},
		{
			Name:  "nil pointer",
			Value: (*int)(nil),
			Want:  true,
		},
		{
			Name:  "non-nil pointer",
			Value: new(int),
			Want:  false,
		},
		{
			Name:  "nil map",
			Value: (map[string]int)(nil),
			Want:  true,
		},
		{
			Name:  "non-nil map",
			Value: map[string]int{},
			Want:  false,
		},
		{
			Name:  "nil slice",
			Value: ([]int)(nil),
			Want:  true,
		},
		{
			Name:  "non-nil slice",
			Value: []int{},
			Want:  false,
		},
		{
			Name:  "nil chan",
			Value: (chan int)(nil),
			Want:  true,
		},
		{
			Name:  "non-nil chan",
			Value: make(chan int),
			Want:  false,
		},
		{
			Name:  "nil func",
			Value: (func())(nil),
			Want:  true,
		},
		{
			Name:  "non-nil func",
			Value: func() {},
			Want:  false,
		},
		{
			Name:  "struct value",
			Value: struct{}{},
			Want:  false,
		},
		{
			Name:  "int value",
			Value: 42,
			Want:  false,
		},
		{
			Name:  "string value",
			Value: "hello",
			Want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			got := IsNil(tc.Value)
			if got != tc.Want {
				t.Fatalf("IsNil() = %v, want %v", got, tc.Want)
			}
		})
	}
}
