package crockford_test

import (
	"testing"

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

			r := testing.Benchmark(func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					dst = crockford.AppendMD5(crockford.Lower, dst, in)
				}
			})
			if r.AllocsPerOp() > 0 {
				t.Errorf("benchmark regression %q: %v", dst, r.MemString())
			}
		})
	}
}
