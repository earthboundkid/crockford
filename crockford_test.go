package crockford_test

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/crockford/v2"
)

func TestMD5(t *testing.T) {
	type testcase struct {
		in   string
		want string
	}
	assert.RunAll(t, map[string]testcase{
		"none":  {"", "tgerspcf02s09tc016cesy22fr"},
		"hello": {"Hello, World!", "cpme4zc8f4m3gcdpcjyrpzratg"},
	}, func(be assert.TB, tc testcase) {
		got := crockford.MD5(crockford.Lower, []byte(tc.in))
		be.Equal(got, tc.want)
	})
}

func TestAppendMD5(t *testing.T) {
	type testcase struct {
		in   string
		want string
	}
	assert.RunAll(t, map[string]testcase{
		"none":  {"", "tgerspcf02s09tc016cesy22fr"},
		"hello": {"Hello, World!", "cpme4zc8f4m3gcdpcjyrpzratg"},
	}, func(be assert.TB, tc testcase) {
		in := []byte(tc.in)
		// plain
		dst := crockford.AppendMD5(crockford.Lower, nil, in)
		be.Equal(tc.want, string(dst))
		// reusing buffer
		dst = crockford.AppendMD5(crockford.Lower, dst[:0], in)
		be.Equal(tc.want, string(dst))
		// appending to buffer
		dst[0] = '*'
		dst = dst[:1]
		dst = crockford.AppendMD5(crockford.Lower, dst, in)
		be.Equal("*"+tc.want, string(dst))

		allocs := testing.AllocsPerRun(100, func() {
			dst = crockford.AppendMD5(crockford.Lower, dst[:0], in)
		})
		be.Falsey(allocs)
	})
}

func TestAppendRandom(t *testing.T) {
	type testcase struct {
		dst []byte
	}
	assert.RunAll(t, map[string]testcase{
		"nil":  {nil},
		"pref": {[]byte("hello ")},
		"cap":  {make([]byte, 0, 8)},
	}, func(be assert.TB, tc testcase) {
		dst := crockford.AppendRandom(crockford.Lower, tc.dst)
		be.SlicesEqual(tc.dst, dst[:len(tc.dst)])
		be.Equal(len(tc.dst)+crockford.LenRandom, len(dst))

		before := string(dst)
		after := string(crockford.AppendRandom(crockford.Lower, tc.dst))
		be.NotEqual(before, after)

		allocs := testing.AllocsPerRun(100, func() {
			dst = crockford.AppendRandom(crockford.Lower, dst[:0])
		})
		be.Falsey(allocs)
	})
}

func TestTime(t *testing.T) {
	type testcase struct {
		want string
	}
	assert.RunAll(t, map[string]testcase{
		"1970-01-01T00:00:00Z": {"00000000"},
		"2000-01-01T12:00:00Z": {"00w6vv20"},
		"2020-01-01T00:00:00Z": {"01f0qr80"},
		"2038-01-19T03:14:07Z": {"01zzzzzz"},
		"2100-01-01T00:00:00Z": {"03t8cnr0"},
	}, func(be assert.TB, tc testcase) {
		name := strings.TrimPrefix(be.Name(), t.Name()+"/")
		when := be.OK(time.Parse("2006-01-02T15:04:05Z", name))
		got := crockford.Time(crockford.Lower, when)
		be.Equal(got, tc.want)
	})
}

func TestAppendTime(t *testing.T) {
	type testcase struct {
		want string
	}
	assert.RunAll(t, map[string]testcase{
		"1970-01-01T00:00:00Z": {"00000000"},
		"2000-01-01T12:00:00Z": {"00w6vv20"},
		"2020-01-01T00:00:00Z": {"01f0qr80"},
		"2038-01-19T03:14:07Z": {"01zzzzzz"},
		"2100-01-01T00:00:00Z": {"03t8cnr0"},
	}, func(be assert.TB, tc testcase) {
		name := strings.TrimPrefix(be.Name(), t.Name()+"/")
		when := be.OK(time.Parse("2006-01-02T15:04:05Z", name))

		dst := crockford.AppendTime(crockford.Lower, when, nil)
		be.Equal(tc.want, string(dst))
		// keep prefixes
		dst = []byte("abc")
		dst = crockford.AppendTime(crockford.Lower, when, dst)
		be.Equal("abc"+tc.want, string(dst))
		// reuse cap
		dst = []byte("12345678--")[:0]
		dst = crockford.AppendTime(crockford.Lower, when, dst)
		dst = dst[:cap(dst)]
		be.Equal(tc.want+"--", string(dst))

		allocs := testing.AllocsPerRun(100, func() {
			dst = crockford.AppendTime(crockford.Lower, when, dst[:0])
		})
		be.Falsey(allocs)
	})
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
		{4, "tgerspcf02s09tc0", "tger-spcf-02s0-9tc0"},
		{4, "cpme4zc8f4m3gcdp", "cpme-4zc8-f4m3-gcdp"},
	} {
		assert.FailsNow(t).
			Run("test", func(be assert.TB) {
				got := crockford.Partition(tc.in, tc.gap)
				be.Equal(tc.out, got)
				be.Equal(simplePartition(tc.in, tc.gap), got)
			}).
			Run("reusing buffer", func(be assert.TB) {
				src := make([]byte, 0, len(tc.out))
				src = append(src, tc.in...)
				b := crockford.AppendPartition(src[:0], src, tc.gap)
				be.Equal(tc.out, string(b))
			}).
			Run("allocations and preserving material before append", func(be assert.TB) {
				src := []byte(tc.in)
				b := make([]byte, len(tc.out)+1)
				b[0] = 'x'
				allocs := testing.AllocsPerRun(100, func() {
					b = b[:1]
					b = crockford.AppendPartition(b, src, tc.gap)
				})
				be.Falsey(allocs)
				be.Equal("x"+tc.out, string(b))
			})
	}
	assert.FailsNow(t).Panicked(func() {
		crockford.Partition("1", -1)
	})
}

func FuzzPartition(f *testing.F) {
	f.Add(1, "")
	f.Add(1, "12")
	f.Add(2, "12")
	f.Add(2, "1234")
	f.Fuzz(func(t *testing.T, gap int, test string) {
		be := assert.FailsNow(t)
		if gap < 1 {
			t.SkipNow()
		}

		s := crockford.Partition(test, gap)
		be.Equal(simplePartition(test, gap), s)
		gaps := len(test) / gap
		if rem := len(test) % gap; rem == 0 && gaps > 0 {
			gaps--
		}
		be.Equal(len(test)+gaps, len(s))
		precount := strings.Count(test, "-")
		postcount := strings.Count(s, "-")
		be.Equal(gaps+precount, postcount)
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

func TestNormalized(t *testing.T) {
	for _, tc := range []struct {
		in, out string
	}{
		{"", ""},
		{"a", "A"},
		{"1IiLl", "11111"},
		{"0Oo", "000"},
		{"-", ""},
		{"AB-C", "ABC"},
		{"AB-C--DEF", "ABCDEF"},
		{"0123456789abcdefghjkmnpqrstvwxyz*~$=u", "0123456789ABCDEFGHJKMNPQRSTVWXYZ*~$=U"},
	} {
		got := crockford.Normalized(tc.in)
		assert.FailsNow(t).Equal(tc.out, got)
	}
}

func TestRandom(t *testing.T) {
	// Reasonably unlikely in a test lol
	a, b := crockford.Random(crockford.Lower), crockford.Random(crockford.Lower)
	assert.FailsNow(t).
		NotEqual(a, b).
		EqualLength(a, crockford.LenRandom).
		EqualLength(b, crockford.LenRandom)
}
