package crockford_test

import (
	"fmt"
	"math"
	"strings"
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
			be.Zero(t, allocs)
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
			be.Zero(t, allocs)
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
			be.Zero(t, allocs)
		})
	}
}

func TestAppend(t *testing.T) {
	for _, tc := range []struct{ in, out string }{
		{"", ""},
		{"\x00", "00"},
		{"\xff\xff\xff\xff\xff", "zzzzzzzz"},
		{"hello world", "d1jprv3f41vpywkccg"},
	} {
		var b []byte
		b = crockford.Append(crockford.Lower, b, []byte(tc.in))
		be.Equal(t, tc.out, string(b))
		b = append(b, '+')
		b = crockford.Append(crockford.Lower, b, []byte(tc.in))
		be.Equal(t, tc.out+"+"+tc.out, string(b))
		src := []byte(tc.in)
		allocs := testing.AllocsPerRun(100, func() {
			b = b[:0]
			b = crockford.Append(crockford.Lower, b, src)
		})
		be.Zero(t, allocs)
	}
}

func ExamplePartition() {
	t := time.Date(1969, 7, 24, 16, 50, 35, 0, time.UTC)
	s := crockford.Time(crockford.Lower, t)
	fmt.Println(crockford.Partition(s, 4))
	// Output:
	// zzzj-satv
}

func TestPartition(t *testing.T) {
	for _, tc := range []struct {
		gap     int
		in, out string
	}{
		{1, "", ""},
		{1, "1", "1"},
		{1, "11", "1-1"},
		{2, "1", "1"},
		{2, "12", "12"},
		{2, "121", "12-1"},
		{2, "1212", "12-12"},
		{2, "12121", "12-12-1"},
		{3, "1231", "123-1"},
		{4, "12341234", "1234-1234"},
	} {
		got := crockford.Partition(tc.in, tc.gap)
		be.Equal(t, tc.out, got)
		be.Equal(t, simplePartition(tc.in, tc.gap), got)

		src := []byte(tc.in)
		b := make([]byte, len(tc.out))
		allocs := testing.AllocsPerRun(100, func() {
			b = b[:0]
			b = crockford.AppendPartition(b, src, tc.gap)
		})
		be.Zero(t, allocs)
		be.Equal(t, tc.out, string(b))
	}
}

func FuzzPartition(f *testing.F) {
	f.Add(1, "")
	f.Add(1, "12")
	f.Add(2, "12")
	f.Add(2, "1234")
	f.Fuzz(func(t *testing.T, gap int, test string) {
		if gap < 1 {
			t.SkipNow()
		}

		s := crockford.Partition(test, gap)
		be.Equal(t, simplePartition(test, gap), s)
		gaps := len(test) / gap
		if rem := len(test) % gap; rem == 0 && gaps > 0 {
			gaps--
		}
		be.Equal(t, len(test)+gaps, len(s))
		precount := strings.Count(test, "-")
		postcount := strings.Count(s, "-")
		be.Equal(t, gaps+precount, postcount)
	})
}

// Same output as partition but it allocates more
func simplePartition(src string, gap int) string {
	return strings.Join(chunk(src, gap), "-")
}

func chunk(s string, size int) []string {
	if len(s) == 0 {
		return nil
	}
	n := int(math.Ceil(float64(len(s)) / float64(size)))
	res := make([]string, 0, n)
	for i := 0; i < n-1; i++ {
		res = append(res, s[i*size:(i+1)*size])
	}
	return append(res, s[(n-1)*size:])
}
