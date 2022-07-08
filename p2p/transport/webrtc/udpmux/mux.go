package udpmux

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	logging "github.com/ipfs/go-log/v2"
	pool "github.com/libp2p/go-buffer-pool"
	"github.com/pion/ice/v2"
	"github.com/pion/stun"
)

var log = logging.Logger("mux")

var (
	errConnNotFound = errors.New("connection not found")
)

const ReceiveMTU = 1500

var _ ice.UDPMux = &udpMux{}

type ufragConnKey struct {
	ufrag  string
	isIPv6 bool
}

// udpMux multiplexes multiple ICE connections over a single net.PacketConn,
// generally a UDP socket.
//
// The connections are indexed by (ufrag, IP address family)
// and by remote address from which the connection has received valid STUN/RTC
// packets.
//
// When a new packet is received on the underlying net.PacketConn, we
// first check the address map to see if there is a connection associated with the
// remote address. If found we forward the packet to the connection. If an associated
// connection is not found, we check to see if the packet is a STUN packet. We then
// fetch the ufrag of the remote from the STUN packet and use it to check if there
// is a connection associated with the (ufrag, IP address family) pair. If found
// we add the association to the address map. If not found, it is a previously
// unseen IP address and the `unknownUfragCallback` callback is invoked.
type udpMux struct {
	socket               net.PacketConn
	unknownUfragCallback func(string, net.Addr) bool

	storage *udpMuxStorage

	// the context controls the lifecycle of the mux
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUDPMux(socket net.PacketConn, unknownUfragCallback func(string, net.Addr) bool) *udpMux {
	ctx, cancel := context.WithCancel(context.Background())
	mux := &udpMux{
		ctx:                  ctx,
		cancel:               cancel,
		socket:               socket,
		unknownUfragCallback: unknownUfragCallback,
		storage:              newUDPMuxStorage(),
	}

	mux.wg.Add(1)
	go func() {
		defer mux.wg.Done()
		mux.readLoop()
	}()
	return mux
}

// GetListenAddresses implements ice.UDPMux
func (mux *udpMux) GetListenAddresses() []net.Addr {
	return []net.Addr{mux.socket.LocalAddr()}
}

// GetConn implements ice.UDPMux
// It creates a net.PacketConn for a given ufrag if an existing
// one cannot be  found. We differentiate IPv4 and IPv6 addresses
// as a remote is capable of being reachable through multiple different
// UDP addresses of the same IP address family (eg. Server-reflexive addresses
// and peer-reflexive addresses).
func (mux *udpMux) GetConn(ufrag string, addr net.Addr) (net.PacketConn, error) {
	a, ok := addr.(*net.UDPAddr)
	if !ok && addr != nil {
		return nil, fmt.Errorf("unexpected address type: %T", addr)
	}
	isIPv6 := ok && a.IP.To4() == nil
	return mux.getOrCreateConn(ufrag, isIPv6, addr)
}

// Close implements ice.UDPMux
func (mux *udpMux) Close() error {
	select {
	case <-mux.ctx.Done():
		return nil
	default:
	}
	mux.cancel()
	mux.socket.Close()
	mux.wg.Wait()
	return nil
}

// RemoveConnByUfrag implements ice.UDPMux
func (mux *udpMux) RemoveConnByUfrag(ufrag string) {
	if ufrag != "" {
		mux.storage.RemoveConnByUfrag(ufrag)
	}
}

func (mux *udpMux) getOrCreateConn(ufrag string, isIPv6 bool, addr net.Addr) (net.PacketConn, error) {
	select {
	case <-mux.ctx.Done():
		return nil, io.ErrClosedPipe
	default:
		conn, _, err := mux.storage.GetOrCreateConn(ufrag, isIPv6, mux, addr)
		return conn, err
	}
}

// writeTo writes a packet to the underlying net.PacketConn
func (mux *udpMux) writeTo(buf []byte, addr net.Addr) (int, error) {
	return mux.socket.WriteTo(buf, addr)
}

func (mux *udpMux) readLoop() {
	for {
		select {
		case <-mux.ctx.Done():
			return
		default:
		}

		buf := pool.Get(ReceiveMTU)

		n, addr, err := mux.socket.ReadFrom(buf)
		if err != nil {
			log.Errorf("error reading from socket: %v", err)
			pool.Put(buf)
			return
		}
		buf = buf[:n]

		// a non-nil error signifies that the packet was not
		// passed on to any connection, and therefore the current
		// function has ownership of the packet. Otherwise, the
		// ownership of the packet is passed to a connection
		if err := mux.processPacket(buf, addr); err != nil {
			pool.Put(buf)
		}
	}
}

func (mux *udpMux) processPacket(buf []byte, addr net.Addr) error {
	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return fmt.Errorf("underlying connection did not return a UDP address")
	}
	isIPv6 := udpAddr.IP.To4() == nil

	// Connections are indexed by remote address. We first
	// check if the remote address has a connection associated
	// with it. If yes, we push the received packet to the connection
	if conn, ok := mux.storage.GetConnByAddr(udpAddr); ok {
		err := conn.Push(buf)
		if err != nil {
			log.Errorf("could not push packet: %v", err)
		}
		return nil
	}

	if !stun.IsMessage(buf) {
		return errConnNotFound
	}

	msg := &stun.Message{Raw: buf}
	if err := msg.Decode(); err != nil || msg.Type != stun.BindingRequest {
		log.Debug("incoming message should be a STUN binding request")
		return err
	}

	ufrag, err := ufragFromSTUNMessage(msg)
	if err != nil {
		log.Debug("could not find STUN username: %w", err)
		return err
	}

	var connCreated bool
	conn, connCreated, err := mux.storage.GetOrCreateConn(ufrag, isIPv6, mux, udpAddr)
	if err != nil {
		log.Debug("could not find create conn: %w", err)
		return err
	}
	if connCreated && mux.unknownUfragCallback != nil {
		if !mux.unknownUfragCallback(ufrag, udpAddr) {
			conn.Close()
			return io.ErrClosedPipe
		}
	}

	if err := conn.Push(buf); err != nil {
		log.Errorf("could not push packet: %v", err)
	}
	return nil
}

// ufragFromSTUNMessage returns the local or ufrag
// from the STUN username attribute. Local ufrag is the ufrag of the
// peer which initiated the connectivity check, e.g in a connectivity
// check from A to B, the username attribute will be B_ufrag:A_ufrag
// with the local ufrag value being A_ufrag. In case of ice-lite, the
// localUfrag value will always be the remote peer's ufrag since ICE-lite
// implementations do not generate connectivity checks. In our specific
// case, since the local and remote ufrag is equal, we can return
// either value.
func ufragFromSTUNMessage(msg *stun.Message) (string, error) {
	attr, err := msg.Get(stun.AttrUsername)
	if err != nil {
		return "", err
	}
	index := bytes.Index(attr, []byte{':'})
	if index == -1 {
		return "", fmt.Errorf("invalid STUN username attribute")
	}
	return string(attr[index+1:]), nil
}

type udpMuxStorage struct {
	sync.Mutex

	ufragMap map[ufragConnKey]*muxedConnection
	addrMap  map[string]*muxedConnection
}

func newUDPMuxStorage() *udpMuxStorage {
	return &udpMuxStorage{
		ufragMap: make(map[ufragConnKey]*muxedConnection),
		addrMap:  make(map[string]*muxedConnection),
	}
}

func (s *udpMuxStorage) RemoveConnByUfrag(ufrag string) {
	s.Lock()
	defer s.Unlock()

	for _, isIPv6 := range [...]bool{true, false} {
		key := ufragConnKey{ufrag: ufrag, isIPv6: isIPv6}
		if conn, ok := s.ufragMap[key]; ok {
			_ = conn.closeConnection()
			delete(s.ufragMap, key)
			delete(s.addrMap, conn.Address().String())
		}
	}
}

func (s *udpMuxStorage) GetConn(ufrag string, isIPv6 bool) (*muxedConnection, bool) {
	key := ufragConnKey{ufrag: ufrag, isIPv6: isIPv6}
	s.Lock()
	conn, ok := s.ufragMap[key]
	s.Unlock()
	return conn, ok
}

func (s *udpMuxStorage) GetOrCreateConn(ufrag string, isIPv6 bool, mux *udpMux, addr net.Addr) (*muxedConnection, bool, error) {
	key := ufragConnKey{ufrag: ufrag, isIPv6: isIPv6}

	s.Lock()
	defer s.Unlock()

	if conn, ok := s.ufragMap[key]; ok {
		return conn, false, nil
	}

	conn := newMuxedConnection(mux, ufrag, addr)
	s.ufragMap[key] = conn
	s.addrMap[addr.String()] = conn

	return conn, true, nil
}

func (s *udpMuxStorage) GetConnByAddr(addr *net.UDPAddr) (*muxedConnection, bool) {
	s.Lock()
	conn, ok := s.addrMap[addr.String()]
	s.Unlock()
	return conn, ok
}
