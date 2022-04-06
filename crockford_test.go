package crockford_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/carlmjohnson/crockford"
)

func EqBytes(t *testing.T) func(want string, got []byte) {
	return func(want string, got []byte) {
		t.Helper()
		if want != string(got) {
			t.Fatalf("want %q; got %q", want, got)
		}
	}
}

func TestEnsure(t *testing.T) {
	cases := map[string]struct {
		size int
		b    []byte
	}{
		"0-nil":    {0, nil},
		"4-nil":    {4, nil},
		"0-sliced": {4, []byte("1234")[:0]},
		"2-sliced": {2, []byte("1234")[:2]},
		"overflow": {2, []byte("1234")},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ret, tar := crockford.Ensure(tc.size, tc.b)
			if len(tar) != tc.size {
				t.Fatalf("bad target: %q", tar)
			}
			if len(ret) != len(tc.b)+tc.size {
				t.Fatalf("bad return: %q", ret)
			}
			if cap(tc.b)-len(tc.b) >= tc.size {
				if bytes.ContainsAny(tar, "\x00") {
					t.Fatalf("overwrote existing cap: %q", ret)
				}
			}
			if !bytes.HasPrefix(ret, tc.b) {
				t.Fatalf("lost prefix: %q", ret)
			}
		})
	}
}

func TestMD5(t *testing.T) {
	cases := map[string]struct {
		in   string
		want string
	}{
		"none":  {"", "tgerspcf02s09tc016cesy22fr"},
		"hello": {"Hello, World!", "cpme4zc8f4m3gcdpcjyrpzratg"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := crockford.MD5(crockford.Lower, []byte(tc.in))
			if got != tc.want {
				t.Fatalf("want %q; got %q", tc.want, got)
			}
		})
	}
}

func TestAppendMD5(t *testing.T) {
	cases := map[string]struct {
		in   string
		want string
	}{
		"none":  {"", "tgerspcf02s09tc016cesy22fr"},
		"hello": {"Hello, World!", "cpme4zc8f4m3gcdpcjyrpzratg"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eq := EqBytes(t)
			in := []byte(tc.in)
			// plain
			dst := crockford.AppendMD5(crockford.Lower, nil, in)
			eq(tc.want, dst)
			// reusing buffer
			dst = crockford.AppendMD5(crockford.Lower, dst[:0], in)
			eq(tc.want, dst)
			// appending to buffer
			dst[0] = '*'
			dst = dst[:1]
			dst = crockford.AppendMD5(crockford.Lower, dst, in)
			eq("*"+tc.want, dst)

			allocs := testing.AllocsPerRun(100, func() {
				dst = crockford.AppendMD5(crockford.Lower, dst[:0], in)
			})
			if allocs > 0 {
				t.Errorf("too many allocs %q: %f", dst, allocs)
			}
		})
	}
}

func TestAppendRandom(t *testing.T) {
	cases := map[string]struct {
		dst []byte
	}{
		"nil":  {nil},
		"pref": {[]byte("hello ")},
		"cap":  {make([]byte, 0, 8)},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dst := crockford.AppendRandom(crockford.Lower, tc.dst)
			if !bytes.HasPrefix(dst, tc.dst) {
				t.Fatalf("lost prefix: %q", dst)
			}
			if len(dst) != len(tc.dst)+crockford.LenRandom {
				t.Fatalf("bad length: %q", dst)
			}
			before := string(dst)
			after := string(crockford.AppendRandom(crockford.Lower, tc.dst))
			if before == after {
				t.Fatalf("results not random: %q == %q", before, after)
			}
			allocs := testing.AllocsPerRun(100, func() {
				dst = crockford.AppendRandom(crockford.Lower, dst[:0])
			})
			if allocs > 0 {
				t.Errorf("too many allocs %q: %f", dst, allocs)
			}
		})
	}
}

func TestTime(t *testing.T) {
	cases := map[string]struct {
		want string
	}{
		"1970-01-01T00:00:00Z": {"00000000"},
		"2000-01-01T12:00:00Z": {"00w6vv20"},
		"2020-01-01T00:00:00Z": {"01f0qr80"},
		"2038-01-19T03:14:07Z": {"01zzzzzz"},
		"2100-01-01T00:00:00Z": {"03t8cnr0"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			when, err := time.Parse("2006-01-02T15:04:05Z", name)
			if err != nil {
				t.Fatal(err)
			}
			got := crockford.Time(crockford.Lower, when)
			if got != tc.want {
				t.Fatalf("want %q; got %q", tc.want, got)
			}
		})
	}
}

func TestAppendTime(t *testing.T) {
	cases := map[string]struct {
		want string
	}{
		"1970-01-01T00:00:00Z": {"00000000"},
		"2000-01-01T12:00:00Z": {"00w6vv20"},
		"2020-01-01T00:00:00Z": {"01f0qr80"},
		"2038-01-19T03:14:07Z": {"01zzzzzz"},
		"2100-01-01T00:00:00Z": {"03t8cnr0"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eq := EqBytes(t)
			when, err := time.Parse("2006-01-02T15:04:05Z", name)
			if err != nil {
				t.Fatal(err)
			}
			dst := crockford.AppendTime(crockford.Lower, when, nil)
			eq(tc.want, dst)
			dst = []byte("abc")
			dst = crockford.AppendTime(crockford.Lower, when, dst)
			if !bytes.HasPrefix(dst, []byte("abc")) {
				t.Fatalf("lost prefix %q", dst)
			}
			dst = []byte("12345678--")[:0]
			dst = crockford.AppendTime(crockford.Lower, when, dst)
			dst = dst[:cap(dst)]
			if !bytes.HasSuffix(dst, []byte("--")) {
				t.Fatalf("lost suffix %q", dst)
			}
			allocs := testing.AllocsPerRun(100, func() {
				dst = crockford.AppendTime(crockford.Lower, when, dst[:0])
			})
			if allocs > 0 {
				t.Errorf("too many allocs %q: %f", dst, allocs)
			}
		})
	}
}
