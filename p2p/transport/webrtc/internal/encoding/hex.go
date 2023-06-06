package encoding

// The code in this file is adapted from the Go standard library's hex package.
// As found in https://cs.opensource.google/go/go/+/refs/tags/go1.20.2:src/encoding/hex/hex.go
//
// The reason we adapted the original code is to allow us to deal with interspersed requirements
// while at the same time hex encoding/decoding, without having to do so in two passes.

import (
	"encoding/hex"
	"errors"
)

// EncodeInterspersedHex encodes a byte slice into a string of hex characters,
// separating each encoded byte with a colon (':').
//
// Example: { 0x01, 0x02, 0x03 } -> "01:02:03"
func EncodeInterspersedHex(src []byte) string {
	if len(src) == 0 {
		return ""
	}
	s := hex.EncodeToString(src)
	n := len(s)
	// Determine number of colons
	colons := n / 2
	if n%2 == 0 {
		colons--
	}
	buffer := make([]byte, n+colons)

	for i, j := 0, 0; i < n; i, j = i+2, j+3 {
		copy(buffer[j:j+2], s[i:i+2])
		if j+3 < len(buffer) {
			buffer[j+2] = ':'
		}
	}
	return string(buffer)
}

// DecodeInterspersedHex decodes a byte slice string of hex characters into a byte slice,
// where the hex characters are expected to be separated by a colon (':').
//
// Example: {'0', '1', ':', '0', '2', ':', '0', '3'} -> { 0x01, 0x02, 0x03 }
func DecodeInterspersedHex(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}
	if len(src) < 2 {
		return nil, hex.ErrLength
	}

	dst := make([]byte, (len(src)+1)/3)
	i, j := 0, 1
	for ; j < len(src); j += 3 { // jump one extra byte for the separator (:)
		p := src[j-1]
		q := src[j]
		if j+1 < len(src) && src[j+1] != ':' {
			return nil, errUnexpectedIntersperseHexChar
		}

		a := reverseHexTable[p]
		b := reverseHexTable[q]
		if a > 0x0f {
			return nil, hex.InvalidByteError(p)
		}
		if b > 0x0f {
			return nil, hex.InvalidByteError(q)
		}
		dst[i] = (a << 4) | b
		i++
	}
	if (len(src)+1)%3 != 0 {
		if len(src)%3 == 0 {
			j -= 1
		}
		// Check for invalid char before reporting bad length,
		// since the invalid char (if present) is an earlier problem.
		if reverseHexTable[src[j-1]] > 0x0f {
			return nil, hex.InvalidByteError(src[j-1])
		}
		return nil, hex.ErrLength
	}
	return dst[:i], nil
}

// DecodeInterpersedHexFromASCIIString decodes an ASCII string of hex characters into a byte slice,
// where the hex characters are expected to be separated by a colon (':').
//
// NOTE that this function returns an error in case the input string contains non-ASCII characters.
//
// Example: "01:02:03" -> { 0x01, 0x02, 0x03 }
func DecodeInterpersedHexFromASCIIString(src string) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}
	if len(src) < 2 {
		return nil, hex.ErrLength
	}

	dst := make([]byte, (len(src)+1)/3)
	i, j := 0, 1
	for ; j < len(src); j += 3 { // jump one extra byte for the separator (:)
		p := src[j-1]
		q := src[j]
		if j+1 < len(src) && src[j+1] != ':' {
			return nil, errUnexpectedIntersperseHexChar
		}

		a := reverseHexTable[p]
		b := reverseHexTable[q]
		if a > 0x0f {
			return nil, hex.InvalidByteError(p)
		}
		if b > 0x0f {
			return nil, hex.InvalidByteError(q)
		}
		dst[i] = (a << 4) | b
		i++
	}
	if (len(src)+1)%3 != 0 {
		if len(src)%3 == 0 {
			j -= 1
		}
		// Check for invalid char before reporting bad length,
		// since the invalid char (if present) is an earlier problem.
		if reverseHexTable[src[j-1]] > 0x0f {
			return nil, hex.InvalidByteError(src[j-1])
		}
		return nil, hex.ErrLength
	}
	return dst[:i], nil
}

var (
	errUnexpectedIntersperseHexChar = errors.New("unexpected character in interspersed hex string")
)

const (
	reverseHexTable = "" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\xff\xff\xff\xff\xff\xff" +
		"\xff\x0a\x0b\x0c\x0d\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\x0a\x0b\x0c\x0d\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff"
)
