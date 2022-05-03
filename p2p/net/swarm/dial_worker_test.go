package swarm

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	csms "github.com/libp2p/go-libp2p/p2p/net/conn-security-multistream"
	tptu "github.com/libp2p/go-libp2p/p2p/net/upgrader"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"

	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/sec/insecure"
	"github.com/libp2p/go-libp2p-core/transport"

	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	msmux "github.com/libp2p/go-stream-muxer-multistream"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/stretchr/testify/require"
)

func newIdentity(t *testing.T) (peer.ID, ic.PrivKey) {
	key, _, err := ic.GenerateECDSAKeyPair(rand.Reader)
	require.NoError(t, err)
	id, err := peer.IDFromPrivateKey(key)
	require.NoError(t, err)
	return id, key
}

func makeSwarm(t *testing.T) *Swarm {
	id, key := newIdentity(t)

	ps, err := pstoremem.NewPeerstore()
	require.NoError(t, err)
	ps.AddPubKey(id, key.GetPublic())
	ps.AddPrivKey(id, key)
	t.Cleanup(func() { ps.Close() })

	s, err := NewSwarm(id, ps, WithDialTimeout(time.Second))
	require.NoError(t, err)

	upgrader := makeUpgrader(t, s)

	var tcpOpts []tcp.Option
	tcpOpts = append(tcpOpts, tcp.DisableReuseport())
	tcpTransport, err := tcp.NewTCPTransport(upgrader, nil, tcpOpts...)
	require.NoError(t, err)
	if err := s.AddTransport(tcpTransport); err != nil {
		t.Fatal(err)
	}
	if err := s.Listen(ma.StringCast("/ip4/127.0.0.1/tcp/0")); err != nil {
		t.Fatal(err)
	}

	quicTransport, err := quic.NewTransport(key, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddTransport(quicTransport); err != nil {
		t.Fatal(err)
	}
	if err := s.Listen(ma.StringCast("/ip4/127.0.0.1/udp/0/quic")); err != nil {
		t.Fatal(err)
	}

	return s
}

func makeUpgrader(t *testing.T, n *Swarm) transport.Upgrader {
	id := n.LocalPeer()
	pk := n.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(insecure.ID, insecure.NewWithIdentity(id, pk))

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)
	u, err := tptu.New(secMuxer, stMuxer)
	require.NoError(t, err)
	return u
}

func TestDialWorkerLoopBasic(t *testing.T) {
	s1 := makeSwarm(t)
	s2 := makeSwarm(t)
	defer s1.Close()
	defer s2.Close()

	s1.Peerstore().AddAddrs(s2.LocalPeer(), s2.ListenAddresses(), peerstore.PermanentAddrTTL)

	reqch := make(chan dialRequest)
	resch := make(chan dialResponse)
	worker := newDialWorker(s1, s2.LocalPeer(), reqch)
	go worker.loop()

	var conn *Conn
	reqch <- dialRequest{ctx: context.Background(), resch: resch}
	select {
	case res := <-resch:
		require.NoError(t, res.err)
		conn = res.conn
	case <-time.After(time.Minute):
		t.Fatal("dial didn't complete")
	}

	s, err := conn.NewStream(context.Background())
	require.NoError(t, err)
	s.Close()

	var conn2 *Conn
	reqch <- dialRequest{ctx: context.Background(), resch: resch}
	select {
	case res := <-resch:
		require.NoError(t, res.err)
		conn2 = res.conn
	case <-time.After(time.Minute):
		t.Fatal("dial didn't complete")
	}

	require.Equal(t, conn, conn2)

	close(reqch)
	worker.wg.Wait()
}

func TestDialWorkerLoopConcurrent(t *testing.T) {
	s1 := makeSwarm(t)
	s2 := makeSwarm(t)
	defer s1.Close()
	defer s2.Close()

	s1.Peerstore().AddAddrs(s2.LocalPeer(), s2.ListenAddresses(), peerstore.PermanentAddrTTL)

	reqch := make(chan dialRequest)
	worker := newDialWorker(s1, s2.LocalPeer(), reqch)
	go worker.loop()

	const dials = 100
	var wg sync.WaitGroup
	resch := make(chan dialResponse, dials)
	for i := 0; i < dials; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reschgo := make(chan dialResponse, 1)
			reqch <- dialRequest{ctx: context.Background(), resch: reschgo}
			select {
			case res := <-reschgo:
				resch <- res
			case <-time.After(time.Minute):
				resch <- dialResponse{err: errors.New("timed out!")}
			}
		}()
	}
	wg.Wait()

	for i := 0; i < dials; i++ {
		res := <-resch
		require.NoError(t, res.err)
	}

	t.Log("all concurrent dials done")

	close(reqch)
	worker.wg.Wait()
}

func TestDialWorkerLoopFailure(t *testing.T) {
	s1 := makeSwarm(t)
	defer s1.Close()

	id, _ := newIdentity(t)
	s1.Peerstore().AddAddrs(id, []ma.Multiaddr{ma.StringCast("/ip4/11.0.0.1/tcp/1234"), ma.StringCast("/ip4/11.0.0.1/udp/1234/quic")}, peerstore.PermanentAddrTTL)

	reqch := make(chan dialRequest)
	resch := make(chan dialResponse)
	worker := newDialWorker(s1, id, reqch)
	go worker.loop()

	reqch <- dialRequest{ctx: context.Background(), resch: resch}
	select {
	case res := <-resch:
		require.Error(t, res.err)
	case <-time.After(time.Minute):
		t.Fatal("dial didn't complete")
	}

	close(reqch)
	worker.wg.Wait()
}

func TestDialWorkerLoopConcurrentFailure(t *testing.T) {
	s1 := makeSwarm(t)
	defer s1.Close()

	id, _ := newIdentity(t)
	s1.Peerstore().AddAddrs(id, []ma.Multiaddr{ma.StringCast("/ip4/11.0.0.1/tcp/1234"), ma.StringCast("/ip4/11.0.0.1/udp/1234/quic")}, peerstore.PermanentAddrTTL)

	reqch := make(chan dialRequest)
	worker := newDialWorker(s1, id, reqch)
	go worker.loop()

	const dials = 100
	var errTimeout = errors.New("timed out!")
	var wg sync.WaitGroup
	resch := make(chan dialResponse, dials)
	for i := 0; i < dials; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reschgo := make(chan dialResponse, 1)
			reqch <- dialRequest{ctx: context.Background(), resch: reschgo}

			select {
			case res := <-reschgo:
				resch <- res
			case <-time.After(time.Minute):
				resch <- dialResponse{err: errTimeout}
			}
		}()
	}
	wg.Wait()

	for i := 0; i < dials; i++ {
		res := <-resch
		require.Error(t, res.err)
		if res.err == errTimeout {
			t.Fatal("dial response timed out")
		}
	}

	t.Log("all concurrent dials done")

	close(reqch)
	worker.wg.Wait()
}

func TestDialWorkerLoopConcurrentMix(t *testing.T) {
	s1 := makeSwarm(t)
	s2 := makeSwarm(t)
	defer s1.Close()
	defer s2.Close()

	s1.Peerstore().AddAddrs(s2.LocalPeer(), s2.ListenAddresses(), peerstore.PermanentAddrTTL)
	s1.Peerstore().AddAddrs(s2.LocalPeer(), []ma.Multiaddr{ma.StringCast("/ip4/11.0.0.1/tcp/1234"), ma.StringCast("/ip4/11.0.0.1/udp/1234/quic")}, peerstore.PermanentAddrTTL)

	reqch := make(chan dialRequest)
	worker := newDialWorker(s1, s2.LocalPeer(), reqch)
	go worker.loop()

	const dials = 100
	var wg sync.WaitGroup
	resch := make(chan dialResponse, dials)
	for i := 0; i < dials; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reschgo := make(chan dialResponse, 1)
			reqch <- dialRequest{ctx: context.Background(), resch: reschgo}
			select {
			case res := <-reschgo:
				resch <- res
			case <-time.After(time.Minute):
				resch <- dialResponse{err: errors.New("timed out!")}
			}
		}()
	}
	wg.Wait()

	for i := 0; i < dials; i++ {
		res := <-resch
		require.NoError(t, res.err)
	}

	t.Log("all concurrent dials done")

	close(reqch)
	worker.wg.Wait()
}

func TestDialWorkerLoopConcurrentFailureStress(t *testing.T) {
	s1 := makeSwarm(t)
	defer s1.Close()

	id, _ := newIdentity(t)

	var addrs []ma.Multiaddr
	for i := 0; i < 200; i++ {
		addrs = append(addrs, ma.StringCast(fmt.Sprintf("/ip4/11.0.0.%d/tcp/%d", i%256, 1234+i)))
	}
	s1.Peerstore().AddAddrs(id, addrs, peerstore.PermanentAddrTTL)

	reqch := make(chan dialRequest)
	worker := newDialWorker(s1, id, reqch)
	go worker.loop()

	const dials = 100
	var errTimeout = errors.New("timed out!")
	var wg sync.WaitGroup
	resch := make(chan dialResponse, dials)
	for i := 0; i < dials; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reschgo := make(chan dialResponse, 1)
			reqch <- dialRequest{ctx: context.Background(), resch: reschgo}
			select {
			case res := <-reschgo:
				resch <- res
			case <-time.After(5 * time.Minute):
				resch <- dialResponse{err: errTimeout}
			}
		}()
	}
	wg.Wait()

	for i := 0; i < dials; i++ {
		res := <-resch
		require.Error(t, res.err)
		if res.err == errTimeout {
			t.Fatal("dial response timed out")
		}
	}

	t.Log("all concurrent dials done")

	close(reqch)
	worker.wg.Wait()
}