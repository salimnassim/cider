package cider

import (
	"errors"
	"slices"
	"testing"
)

func TestParserAny(t *testing.T) {

	type tc struct {
		input string
		want  opSet
	}

	tcs := []tc{
		{
			input: "SET key value another arg XX GET EX 5",
			want: opSet{
				key:     "key",
				value:   []byte("value another arg"),
				nx:      false,
				xx:      true,
				get:     true,
				keepttl: false,
			},
		},
		{
			input: `SET key really long really long really long really long really long really long really long really long really long really really long really long really long really long really long really long really long really long really long really long really long really really long really long XX GET EX 5`,
			want: opSet{
				key:     "key",
				value:   []byte(`really long really long really long really long really long really long really long really long really long really really long really long really long really long really long really long really long really long really long really long really long really really long really long`),
				nx:      false,
				xx:      true,
				get:     true,
				keepttl: false,
			},
		},
		{
			input: "SET key value another \"quoted string\" arg XX GET EX 5",
			want: opSet{
				key:     "key",
				value:   []byte("value another \"quoted string\" arg"),
				nx:      false,
				xx:      true,
				get:     true,
				keepttl: false,
			},
		},
		{
			input: "SET key value another arg NX EX 5 KEEPTTL",
			want: opSet{
				key:     "key",
				value:   []byte("value another arg"),
				nx:      true,
				xx:      false,
				get:     false,
				keepttl: true,
			},
		},
	}

	for _, tc := range tcs {
		v, err := ParseCommandAny([]byte(tc.input))
		if err != nil {
			t.Error(err)
		}
		op := v.(opSet)

		if tc.want.key != op.key {
			t.Errorf("want: %v, got %v", tc.want.key, op.key)
		}

		if slices.Compare(tc.want.value, op.value) != 0 {
			t.Errorf("want: %s,\n got %s", tc.want.value, op.value)
		}

		if tc.want.nx != op.nx {
			t.Errorf("want: %v, got %v", tc.want.nx, op.nx)
		}

		if tc.want.xx != op.xx {
			t.Errorf("want: %v, got %v", tc.want.xx, op.xx)
		}

		if tc.want.keepttl != op.keepttl {
			t.Errorf("want: %v, got %v", tc.want.xx, op.xx)
		}
	}

}

func TestParserAnyGet(t *testing.T) {

	type tc struct {
		input string
		want  opGet
	}

	tcs := []tc{
		{
			input: "GET key",
			want: opGet{
				key: "key",
			},
		},
		{
			input: "GET key not key foo baz",
			want: opGet{
				key: "key",
			},
		},
	}

	for _, tc := range tcs {
		v, err := ParseCommandAny([]byte(tc.input))
		if err != nil {
			t.Error(err)
		}
		op := v.(opGet)

		if tc.want.key != op.key {
			t.Errorf("want: %v, got %v", tc.want.key, op.key)
		}
	}
}

func TestParserAnyErrors(t *testing.T) {

	type tc struct {
		input     string
		want      opSet
		wantError error
	}

	// tests for arg conflict handling, should error
	tce := []tc{
		{
			input: "SET key value NX XX GET EX 5",
			want: opSet{
				key:   "key",
				value: []byte("value another arg"),
			},
			wantError: errors.New("NX already set in this command"),
		},
		{
			input: "SET key value XX NX GET EX 5",
			want: opSet{
				key:   "key",
				value: []byte("value another arg"),
			},
			wantError: errors.New("XX already set in this command"),
		},
		{
			input: "SET key value GET EX 5 EXAT 500",
			want: opSet{
				key:   "key",
				value: []byte("value another arg"),
			},
			wantError: errors.New("EX already set in this command"),
		},
		{
			input: "SET key value GET EXAT 500 EX 5",
			want: opSet{
				key:   "key",
				value: []byte("value another arg"),
			},
			wantError: errors.New("EXAT already set in this command"),
		},
		{
			input: "SET key value GET EXAT",
			want: opSet{
				key:   "key",
				value: []byte("value another arg"),
			},
			wantError: errors.New("EXAT value missing"),
		},
		{
			input: "SET key value GET EX",
			want: opSet{
				key:   "key",
				value: []byte("value another arg"),
			},
			wantError: errors.New("EX value missing"),
		},
	}

	for _, tc := range tce {
		_, err := ParseCommandAny([]byte(tc.input))
		if err != nil {
			if err.Error() != tc.wantError.Error() {
				t.Errorf("got: %s, want %s (input: %s)", err, tc.wantError, tc.input)
			}
		}
	}
}

func TestOperations(t *testing.T) {
	type tc struct {
		input string
		want  operation
	}

	tcs := []tc{
		{
			input: "SET foo baz",
			want: operation{
				name:  "SET",
				keys:  []string{"foo"},
				value: "baz",
			},
		},
		{
			input: "SET foo 1",
			want: operation{
				name:  "SET",
				keys:  []string{"foo"},
				value: "1",
			},
		},
		{
			input: `SET foo "quoted string with spaces"`,
			want: operation{
				name:  "SET",
				keys:  []string{"foo"},
				value: `"quoted string with spaces"`,
			},
		},
		{
			input: "GET",
			want: operation{
				name:  "GET",
				keys:  []string{""},
				value: "",
			},
		},
		{
			input: "GET foo",
			want: operation{
				name:  "GET",
				keys:  []string{"foo"},
				value: "",
			},
		},
		{
			input: "DEL foo baz qax",
			want: operation{
				name:  "DEL",
				keys:  []string{"foo", "baz", "qax"},
				value: "",
			},
		},
		{
			input: "EXISTS foo baz quu",
			want: operation{
				name:  "EXISTS",
				keys:  []string{"foo", "baz", "quu"},
				value: "",
			},
		},
		{
			input: "EXPIRE foo 500",
			want: operation{
				name:  "EXPIRE",
				keys:  []string{"foo"},
				value: "500",
			},
		},
	}

	for _, v := range tcs {
		op, err := ParseCommand([]byte(v.input))
		if err != nil {
			continue
		}
		if op.name != v.want.name {
			t.Errorf("op name does not match, want: %s, got %s", v.want.name, op.name)
		}
		if slices.Compare(op.keys, v.want.keys) != 0 {
			t.Errorf("op keys dont not match, want: %v, got %v", v.want.keys, op.keys)
		}
		if op.value != v.want.value {
			t.Errorf("op value does not match, want: %v, got %v", v.want.value, op.value)
		}
	}

}
