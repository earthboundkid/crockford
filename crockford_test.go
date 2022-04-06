package crockford_test

import (
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/crockford"
)

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
			be.Equal(t, tc.want, got)
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
			in := []byte(tc.in)
			// plain
			dst := crockford.AppendMD5(crockford.Lower, nil, in)
			be.Equal(t, tc.want, string(dst))
			// reusing buffer
			dst = crockford.AppendMD5(crockford.Lower, dst[:0], in)
			be.Equal(t, tc.want, string(dst))
			// appending to buffer
			dst[0] = '*'
			dst = dst[:1]
			dst = crockford.AppendMD5(crockford.Lower, dst, in)
			be.Equal(t, "*"+tc.want, string(dst))

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
			be.AllEqual(t, tc.dst, dst[:len(tc.dst)])
			be.Equal(t, len(tc.dst)+crockford.LenRandom, len(dst))

			before := string(dst)
			after := string(crockford.AppendRandom(crockford.Lower, tc.dst))
			be.Unequal(t, before, after)

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
			be.NilErr(t, err)
			got := crockford.Time(crockford.Lower, when)
			be.Equal(t, tc.want, got)
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
			when, err := time.Parse("2006-01-02T15:04:05Z", name)
			be.NilErr(t, err)

			dst := crockford.AppendTime(crockford.Lower, when, nil)
			be.Equal(t, tc.want, string(dst))
			// keep prefixes
			dst = []byte("abc")
			dst = crockford.AppendTime(crockford.Lower, when, dst)
			be.Equal(t, "abc"+tc.want, string(dst))
			// reuse cap
			dst = []byte("12345678--")[:0]
			dst = crockford.AppendTime(crockford.Lower, when, dst)
			dst = dst[:cap(dst)]
			be.Equal(t, tc.want+"--", string(dst))

			allocs := testing.AllocsPerRun(100, func() {
				dst = crockford.AppendTime(crockford.Lower, when, dst[:0])
			})
			if allocs > 0 {
				t.Errorf("too many allocs %q: %f", dst, allocs)
			}
		})
	}
}
