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
			input: "SET key value XX",
			want: opSet{
				key:   "key",
				value: []byte("value"),
				xx:    true,
			},
		},
		{
			input: "SET key value another arg XX GET EX 42",
			want: opSet{
				key:     "key",
				value:   []byte("value another arg"),
				nx:      false,
				xx:      true,
				get:     true,
				keepttl: false,
				ex:      42,
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
				ex:      5,
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
				ex:      5,
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

func TestParserAnyDel(t *testing.T) {
	type tc struct {
		input string
		want  opDel
	}

	tcs := []tc{
		{
			input: "DEL one two three",
			want: opDel{
				keys: []string{"one", "two", "three"},
			},
		},
		{
			input: "DEL one",
			want: opDel{
				keys: []string{"one"},
			},
		},
		{
			input: "DEL one two three 12345",
			want: opDel{
				keys: []string{"one", "two", "three", "12345"},
			},
		},
	}

	for _, tc := range tcs {
		v, err := ParseCommandAny([]byte(tc.input))
		if err != nil {
			t.Error(err)
		}
		op := v.(opDel)

		if slices.Compare(op.keys, tc.want.keys) != 0 {
			t.Errorf("got: %v, want: %v", op.keys, tc.want.keys)
		}
	}
}

func TestParserAnyExpire(t *testing.T) {
	type tc struct {
		input string
		want  opExpire
	}

	tcs := []tc{
		{
			input: "EXPIRE foo 42 NX",
			want: opExpire{
				key: "foo",
				ttl: 42,
				nx:  true,
			},
		},
		{
			input: "EXPIRE foo 42",
			want: opExpire{
				key: "foo",
				ttl: 42,
			},
		},
		{
			input: "EXPIRE foo 101 GT",
			want: opExpire{
				key: "foo",
				ttl: 101,
				gt:  true,
			},
		},
	}

	for _, tc := range tcs {
		v, err := ParseCommandAny([]byte(tc.input))
		if err != nil {
			t.Error(err)
		}
		op := v.(opExpire)

		if op.key != tc.want.key {
			t.Errorf("got: %v, want: %v", op.key, tc.want.key)
		}

		if op.ttl != tc.want.ttl {
			t.Errorf("got: %v, want: %v", op.ttl, tc.want.ttl)
		}

		if op.gt != tc.want.gt {
			t.Errorf("got: %v, want: %v", op.gt, tc.want.gt)
		}

		if op.lt != tc.want.lt {
			t.Errorf("got: %v, want: %v", op.lt, tc.want.lt)
		}

		if op.nx != tc.want.nx {
			t.Errorf("got: %v, want: %v", op.nx, tc.want.nx)
		}

		if op.xx != tc.want.xx {
			t.Errorf("got: %v, want: %v", op.nx, tc.want.nx)
		}

	}
}

func TestParserAnyExists(t *testing.T) {
	type tc struct {
		input string
		want  opExists
	}

	tcs := []tc{
		{
			input: "EXISTS one two three",
			want: opExists{
				keys: []string{"one", "two", "three"},
			},
		},
		{
			input: "EXISTS one",
			want: opExists{
				keys: []string{"one"},
			},
		},
		{
			input: "EXISTS one two three 12345",
			want: opExists{
				keys: []string{"one", "two", "three", "12345"},
			},
		},
	}

	for _, tc := range tcs {
		v, err := ParseCommandAny([]byte(tc.input))
		if err != nil {
			t.Error(err)
		}
		op := v.(opExists)

		if slices.Compare(op.keys, tc.want.keys) != 0 {
			t.Errorf("got: %v, want: %v", op.keys, tc.want.keys)
		}
	}
}

func TestParerAnyIncr(t *testing.T) {
	type tc struct {
		input string
		want  opIncr
	}

	tcs := []tc{
		{
			input: "INCR foo",
			want: opIncr{
				key: "foo",
			},
		},
		{
			input: "INCR baz",
			want: opIncr{
				key: "baz",
			},
		},
	}

	for _, tc := range tcs {
		v, err := ParseCommandAny([]byte(tc.input))
		if err != nil {
			t.Error(err)
		}

		op := v.(opIncr)

		if op.key != tc.want.key {
			t.Errorf("got %v, want %v", op.key, tc.want.key)
		}
	}
}

func TestParerAnyDecr(t *testing.T) {
	type tc struct {
		input string
		want  opDecr
	}

	tcs := []tc{
		{
			input: "DECR foo",
			want: opDecr{
				key: "foo",
			},
		},
		{
			input: "DECR baz",
			want: opDecr{
				key: "baz",
			},
		},
	}

	for _, tc := range tcs {
		v, err := ParseCommandAny([]byte(tc.input))
		if err != nil {
			t.Error(err)
		}

		op := v.(opDecr)

		if op.key != tc.want.key {
			t.Errorf("got %v, want %v", op.key, tc.want.key)
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
