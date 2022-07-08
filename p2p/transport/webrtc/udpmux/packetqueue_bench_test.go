package udpmux

import (
	"context"
	"fmt"
	"testing"

	pool "github.com/libp2p/go-buffer-pool"
)

var sizes = []int{
	1,
	10,
	100,
	maxPacketsInQueue,
}

func BenchmarkQueue(b *testing.B) {
	ctx := context.Background()
	for _, dequeue := range [...]bool{true, false} {
		for _, input := range sizes {
			testCase := fmt.Sprintf("enqueue_%d", input)
			if dequeue {
				testCase = testCase + "_dequeue"
			}
			b.Run(testCase, func(b *testing.B) {
				pq := newPacketQueue()
				buf := make([]byte, 256)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for k := 0; k < input; k++ {
						pq.Push(ctx, pool.Get(255))
					}
					for k := 0; k < input; k++ {
						pq.Pop(ctx, buf)
					}
				}
			})
		}
	}
}
