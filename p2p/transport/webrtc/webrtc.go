// libp2pwebrtc implements the WebRTC transport for go-libp2p,
// as officially described in https://github.com/libp2p/specs/tree/cfcf0230b2f5f11ed6dd060f97305faa973abed2/webrtc.
//
// Benchmarks on how this transport compares to other transports can be found in
// https://github.com/libp2p/libp2p-go-webrtc-benchmarks.
//
// Entrypoint for the logic of this Transport can be found in `transport.go`, where the WebRTC transport is implemented,
// used both by the client for Dialing as well as the server for Listening. Starting from there you should be able to follow
// the logic from start to finish.
//
// In the udpmux subpackage you can find the logic for multiplexing multiple WebRTC (ICE) connections over a single UDP socket.
//
// The pb subpackage contains the protobuf definitions for the signaling protocol used by this transport,
// which is taken verbatim from the "Multiplexing" chapter of the WebRTC spec.
package libp2pwebrtc
