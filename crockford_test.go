package crockford_test

import (
	"bytes"
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
					dst = dst[:0]
					dst = crockford.AppendMD5(crockford.Lower, dst, in)
				}
			})
			if r.AllocsPerOp() != 0 {
				t.Errorf("benchmark regression %q: %v", dst, r.MemString())
			}
		})
	}
}
