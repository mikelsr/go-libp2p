package udpmux

import (
	"context"
	"errors"
	"net"
	"time"
)

var _ net.PacketConn = &muxedConnection{}

var errAlreadyClosed = errors.New("already closed")

// muxedConnection provides a net.PacketConn abstraction
// over packetQueue and adds the ability to store addresses
// from which this connection (indexed by ufrag) received
// data.
type muxedConnection struct {
	ctx    context.Context
	cancel context.CancelFunc
	pq     *packetQueue
	ufrag  string
	addr   net.Addr
	mux    *udpMux
}

var _ net.PacketConn = (*muxedConnection)(nil)

func newMuxedConnection(mux *udpMux, ufrag string, addr net.Addr) *muxedConnection {
	ctx, cancel := context.WithCancel(mux.ctx)
	return &muxedConnection{
		ctx:    ctx,
		cancel: cancel,
		pq:     newPacketQueue(),
		ufrag:  ufrag,
		addr:   addr,
		mux:    mux,
	}
}

func (conn *muxedConnection) Push(buf []byte) error {
	return conn.pq.Push(conn.ctx, buf)
}

// Close implements net.PacketConn
func (conn *muxedConnection) Close() error {
	_ = conn.closeConnection()
	conn.mux.RemoveConnByUfrag(conn.ufrag)
	return nil
}

// LocalAddr implements net.PacketConn
func (conn *muxedConnection) LocalAddr() net.Addr {
	return conn.mux.socket.LocalAddr()
}

func (conn *muxedConnection) Address() net.Addr {
	return conn.addr
}

// ReadFrom implements net.PacketConn
func (conn *muxedConnection) ReadFrom(p []byte) (int, net.Addr, error) {
	n, err := conn.pq.Pop(conn.ctx, p)
	return n, conn.addr, err
}

// SetDeadline implements net.PacketConn
func (*muxedConnection) SetDeadline(t time.Time) error {
	// no deadline is desired here
	return nil
}

// SetReadDeadline implements net.PacketConn
func (*muxedConnection) SetReadDeadline(t time.Time) error {
	// no read deadline is desired here
	return nil
}

// SetWriteDeadline implements net.PacketConn
func (*muxedConnection) SetWriteDeadline(t time.Time) error {
	// no write deadline is desired here
	return nil
}

// WriteTo implements net.PacketConn
func (conn *muxedConnection) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return conn.mux.writeTo(p, addr)
}

func (conn *muxedConnection) closeConnection() error {
	select {
	case <-conn.ctx.Done():
		return errAlreadyClosed
	default:
	}
	conn.cancel()
	return nil
}
