// Crockford implements the Crockford base 32 encoding
//
// See https://www.crockford.com/base32.html
package crockford

import (
	"bytes"
	"encoding/base32"
	"time"
)

const (
	LowercaseAlphabet = "0123456789abcdefghjkmnpqrstvwxyz"
	UppercaseAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	UppercaseChecksum = UppercaseAlphabet + "*~$=U"
	LowercaseChecksum = LowercaseAlphabet + "*~$=u"
)

var (
	Lower = base32.NewEncoding(LowercaseAlphabet).WithPadding(base32.NoPadding)
	Upper = base32.NewEncoding(UppercaseAlphabet).WithPadding(base32.NoPadding)
)

// Time encodes the Unix time as a 40-bit number
func Time(e base32.Encoding, t time.Time) []byte {
	ut := t.Unix()
	var (
		src [5]byte
		dst [8]byte
	)
	src[0] = byte(ut >> 32)
	src[1] = byte(ut >> 24)
	src[2] = byte(ut >> 16)
	src[3] = byte(ut >> 8)
	src[4] = byte(ut)
	e.Encode(dst[:], src[:])
	return dst[:]
}

// mod calculates the big endian modulus of the byte string
func mod(b []byte, m int) (rem int) {
	for _, c := range b {
		rem = (rem*1<<8 + int(c)) % m
	}
	return
}

func Checksum(src []byte, uppercase bool) byte {
	alphabet := LowercaseChecksum
	if uppercase {
		alphabet = UppercaseChecksum
	}
	return alphabet[mod(src, 37)]
}

func normUpper(r rune) rune {
	switch r {
	case '0', 'O', 'o':
		return '0'
	case '1', 'I', 'i':
		return '1'
	case '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'V', 'W', 'X', 'Y', 'Z':
		return r
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'k', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z':
		return r + 'A' - 'a'
	}
	return -1
}

func Normalize(src []byte) []byte {
	return bytes.Map(normUpper, src)
}

func normLower(r rune) rune {
	switch r {
	case '0', 'O', 'o':
		return '0'
	case '1', 'I', 'i':
		return '1'
	case '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'j', 'k', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z':
		return r
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'V', 'W', 'X', 'Y', 'Z':
		return r + 'a' - 'A'
	}
	return -1

}

func NormalizeLower(src []byte) []byte {
	return bytes.Map(normUpper, src)
}
