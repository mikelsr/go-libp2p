package udpmux

import (
	"context"
	"testing"
	"time"

	pool "github.com/libp2p/go-buffer-pool"
	"github.com/stretchr/testify/require"
)

func TestPacketQueue_QueuePacketsForRead(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pq := newPacketQueue()
	pq.Push(ctx, []byte{1, 2, 3})
	pq.Push(ctx, []byte{5, 6, 7, 8})

	buf := pool.Get(100)
	size, err := pq.Pop(ctx, buf)
	require.NoError(t, err)
	require.Equal(t, size, 3)

	size, err = pq.Pop(ctx, buf)
	require.NoError(t, err)
	require.Equal(t, size, 4)
}

func TestPacketQueue_WaitsForData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pq := newPacketQueue()
	buf := pool.Get(100)

	timer := time.AfterFunc(200*time.Millisecond, func() {
		pq.Push(ctx, []byte{5, 6, 7, 8})
	})

	defer timer.Stop()
	size, err := pq.Pop(ctx, buf)
	require.NoError(t, err)
	require.Equal(t, size, 4)
}

func TestPacketQueue_DropsPacketsWhenQueueIsFull(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pq := newPacketQueue()
	for i := 0; i < maxPacketsInQueue; i++ {
		buf := pool.Get(255)
		err := pq.Push(ctx, buf)
		require.NoError(t, err)
	}

	buf := pool.Get(255)
	err := pq.Push(ctx, buf)
	require.ErrorIs(t, err, errTooManyPackets)
}
