package internal

import (
	"github.com/libp2p/go-libp2p/p2p/transport/webrtc/internal/encoding"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multibase"
	mh "github.com/multiformats/go-multihash"

	"github.com/pion/webrtc/v3"
)

func DecodeRemoteFingerprint(maddr ma.Multiaddr) (*mh.DecodedMultihash, error) {
	remoteFingerprintMultibase, err := maddr.ValueForProtocol(ma.P_CERTHASH)
	if err != nil {
		return nil, err
	}
	_, data, err := multibase.Decode(remoteFingerprintMultibase)
	if err != nil {
		return nil, err
	}
	return mh.Decode(data)
}

func EncodeDTLSFingerprint(fp webrtc.DTLSFingerprint) (string, error) {
	digest, err := encoding.DecodeInterpersedHexFromASCIIString(fp.Value)
	if err != nil {
		return "", err
	}
	encoded, err := mh.Encode(digest, mh.SHA2_256)
	if err != nil {
		return "", err
	}
	return multibase.Encode(multibase.Base64url, encoded)
}
