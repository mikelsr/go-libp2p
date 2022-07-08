package udpmux

import (
	"context"
	"errors"
	"sync"

	pool "github.com/libp2p/go-buffer-pool"
)

type packet struct {
	buf []byte
}

var (
	errTooManyPackets    = errors.New("too many packets in queue; dropping")
	errEmptyPacketQueue  = errors.New("packet queue is empty")
	errPacketQueueClosed = errors.New("packet queue closed")
)

const maxPacketsInQueue = 128

type packetQueue struct {
	packetsMux sync.Mutex
	packetsCh  chan struct{}
	packets    []packet
}

func newPacketQueue() *packetQueue {
	return &packetQueue{
		packetsCh: make(chan struct{}, 1),
	}
}

// Pop reads a packet from the packetQueue or blocks until
// either a packet becomes available or the queue is closed.
func (pq *packetQueue) Pop(ctx context.Context, buf []byte) (int, error) {
	select {
	case <-pq.packetsCh:
		pq.packetsMux.Lock()
		defer pq.packetsMux.Unlock()

		if len(pq.packets) == 0 {
			return 0, errEmptyPacketQueue
		}
		p := pq.packets[0]

		n := copy(buf, p.buf)
		if n == len(p.buf) {
			// only move packet from queue if we read all
			pq.packets = pq.packets[1:]
			pool.Put(p.buf)
		} else {
			// otherwise we need to keep the packet in the queue
			// but do update the buf
			pq.packets[0].buf = p.buf[n:]
		}

		if len(pq.packets) > 0 {
			// to make sure a next pop call will work
			select {
			case pq.packetsCh <- struct{}{}:
			default:
			}
		}

		return n, nil

	// It is desired to allow reads of this channel even
	// when pq.ctx.Done() is already closed.
	case <-ctx.Done():
		return 0, errPacketQueueClosed
	}
}

// Push adds a packet to the packetQueue
func (pq *packetQueue) Push(ctx context.Context, buf []byte) error {
	pq.packetsMux.Lock()
	defer pq.packetsMux.Unlock()

	if len(pq.packets) >= maxPacketsInQueue {
		return errTooManyPackets
	}

	pq.packets = append(pq.packets, packet{buf})
	select {
	case pq.packetsCh <- struct{}{}:
	default:
	}

	return nil
}
