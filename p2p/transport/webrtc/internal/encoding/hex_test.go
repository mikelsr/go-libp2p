package encoding

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeInterspersedHex(t *testing.T) {
	b, err := hex.DecodeString("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	require.NoError(t, err)
	require.Equal(t, "ba:78:16:bf:8f:01:cf:ea:41:41:40:de:5d:ae:22:23:b0:03:61:a3:96:17:7a:9c:b4:10:ff:61:f2:00:15:ad", EncodeInterspersedHex(b))
}

func TestEncodeInterspersedHexToBuilder(t *testing.T) {
	b, err := hex.DecodeString("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	require.NoError(t, err)
	var builder strings.Builder
	EncodeInterspersedHexToBuilder(b, &builder)
	require.Equal(t, "ba:78:16:bf:8f:01:cf:ea:41:41:40:de:5d:ae:22:23:b0:03:61:a3:96:17:7a:9c:b4:10:ff:61:f2:00:15:ad", builder.String())
}

func TestDecodeInterpersedHexStringLowerCase(t *testing.T) {
	b, err := DecodeInterpersedHexFromASCIIString("ba:78:16:bf:8f:01:cf:ea:41:41:40:de:5d:ae:22:23:b0:03:61:a3:96:17:7a:9c:b4:10:ff:61:f2:00:15:ad")
	require.NoError(t, err)
	require.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", hex.EncodeToString(b))
}

func TestDecodeInterpersedHexStringMixedCase(t *testing.T) {
	b, err := DecodeInterpersedHexFromASCIIString("Ba:78:16:BF:8F:01:cf:ea:41:41:40:De:5d:ae:22:23:b0:03:61:a3:96:17:7a:9c:b4:10:FF:61:f2:00:15:ad")
	require.NoError(t, err)
	require.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", hex.EncodeToString(b))
}

func TestDecodeInterpersedHexStringOneByte(t *testing.T) {
	b, err := DecodeInterpersedHexFromASCIIString("ba")
	require.NoError(t, err)
	require.Equal(t, "ba", hex.EncodeToString(b))
}

func TestDecodeInterpersedHexBytesLowerCase(t *testing.T) {
	b, err := DecodeInterspersedHex([]byte("ba:78:16:bf:8f:01:cf:ea:41:41:40:de:5d:ae:22:23:b0:03:61:a3:96:17:7a:9c:b4:10:ff:61:f2:00:15:ad"))
	require.NoError(t, err)
	require.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", hex.EncodeToString(b))
}

func TestDecodeInterpersedHexBytesMixedCase(t *testing.T) {
	b, err := DecodeInterspersedHex([]byte("Ba:78:16:BF:8F:01:cf:ea:41:41:40:De:5d:ae:22:23:b0:03:61:a3:96:17:7a:9c:b4:10:FF:61:f2:00:15:ad"))
	require.NoError(t, err)
	require.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", hex.EncodeToString(b))
}

func TestDecodeInterpersedHexBytesOneByte(t *testing.T) {
	b, err := DecodeInterspersedHex([]byte("ba"))
	require.NoError(t, err)
	require.Equal(t, "ba", hex.EncodeToString(b))
}

func TestEncodeInterperseHexNilSlice(t *testing.T) {
	require.Equal(t, "", EncodeInterspersedHex(nil))
	require.Equal(t, "", EncodeInterspersedHex([]byte{}))
}

func TestDecodeInterspersedHexNilSlice(t *testing.T) {
	for _, v := range [][]byte{nil, {}} {
		b, err := DecodeInterspersedHex(v)
		require.NoError(t, err)
		require.Equal(t, []byte{}, b)
	}
}

func TestDecodeInterpersedHexFromASCIIStringEmpty(t *testing.T) {
	b, err := DecodeInterpersedHexFromASCIIString("")
	require.NoError(t, err)
	require.Equal(t, []byte{}, b)
}

func TestDecodeInterpersedHexInvalid(t *testing.T) {
	for _, v := range []string{"0", "0000", "000"} {
		_, err := DecodeInterspersedHex([]byte(v))
		require.Error(t, err)
	}
}

func TestDecodeInterpersedHexValid(t *testing.T) {
	b, err := DecodeInterspersedHex([]byte("00"))
	require.NoError(t, err)
	require.Equal(t, []byte{0}, b)
}

func TestDecodeInterpersedHexFromASCIIStringInvalid(t *testing.T) {
	for _, v := range []string{"0", "0000", "000"} {
		_, err := DecodeInterpersedHexFromASCIIString(v)
		require.Error(t, err)
	}
}

func TestDecodeInterpersedHexFromASCIIStringValid(t *testing.T) {
	b, err := DecodeInterpersedHexFromASCIIString("00")
	require.NoError(t, err)
	require.Equal(t, []byte{0}, b)
}

func FuzzInterpersedHex(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		decoded, err := DecodeInterspersedHex(b)
		if err != nil {
			return
		}
		encoded := EncodeInterspersedHex(decoded)
		require.Equal(t, strings.ToLower(string(b)), encoded)
	})
}

func FuzzInterspersedHexASCII(f *testing.F) {
	f.Fuzz(func(t *testing.T, s string) {
		decoded, err := DecodeInterpersedHexFromASCIIString(s)
		if err != nil {
			return
		}
		encoded := EncodeInterspersedHex(decoded)
		require.Equal(t, strings.ToLower(s), encoded)
	})
}
