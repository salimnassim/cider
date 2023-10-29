package cider

import (
	"slices"
	"testing"
)

func TestOperations(t *testing.T) {

	type tc struct {
		input string
		want  Operation
	}

	tcs := []tc{
		{
			input: "SET foo baz",
			want: Operation{
				Name:  "SET",
				Keys:  []string{"foo"},
				Value: "baz",
			},
		},
		{
			input: "GET foo",
			want: Operation{
				Name:  "GET",
				Keys:  []string{"foo"},
				Value: "",
			},
		},
		{
			input: "DEL foo baz qax",
			want: Operation{
				Name:  "DEL",
				Keys:  []string{"foo", "baz", "qax"},
				Value: "",
			},
		},
		{
			input: "EXISTS foo baz quu",
			want: Operation{
				Name:  "EXISTS",
				Keys:  []string{"foo", "baz", "quu"},
				Value: "",
			},
		},
	}

	for _, v := range tcs {
		op := ParseCommand([]byte(v.input))
		if op.Name != v.want.Name {
			t.Errorf("op name does not match, want: %s, got %s", v.want.Name, op.Name)
		}
		if slices.Compare(op.Keys, v.want.Keys) != 0 {
			t.Errorf("op keys dont not match, want: %v, got %v", v.want.Keys, op.Keys)
		}
		if op.Value != v.want.Value {
			t.Errorf("op value does not match, want: %v, got %v", v.want.Value, op.Value)
		}
	}

}
