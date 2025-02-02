package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/mikelsr/go-libp2p"
	"github.com/mikelsr/go-libp2p/core/peer"
	"github.com/mikelsr/go-libp2p/p2p/protocol/ping"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	rcmgr "github.com/mikelsr/go-libp2p/p2p/host/resource-manager"
	rcmgrObs "github.com/mikelsr/go-libp2p/p2p/host/resource-manager/obs"
)

const ClientCount = 32

func main() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":2112", nil)
	}()

	rcmgrObs.MustRegisterWith(prometheus.DefaultRegisterer)

	str, err := rcmgrObs.NewStatsTraceReporter()
	if err != nil {
		log.Fatal(err)
	}

	rmgr, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.DefaultLimits.AutoScale()), rcmgr.WithTraceReporter(str))
	if err != nil {
		log.Fatal(err)
	}
	server, err := libp2p.New(libp2p.ResourceManager(rmgr))
	if err != nil {
		log.Fatal(err)
	}

	// Make a bunch of clients that all ping the server at various times
	wg := sync.WaitGroup{}
	for i := 0; i < ClientCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			time.Sleep(time.Duration(i%100) * 100 * time.Millisecond)
			newClient(peer.AddrInfo{
				ID:    server.ID(),
				Addrs: server.Addrs(),
			}, i)
		}(i)
	}
	wg.Wait()
}

func newClient(serverInfo peer.AddrInfo, pings int) {
	// Sleep some random amount of time to spread out the clients so the graphs look more interesting
	time.Sleep(time.Duration(rand.Intn(100)) * time.Second)
	fmt.Println("Started client", pings)

	client, err := libp2p.New(
		// We just want metrics from the server
		libp2p.DisableMetrics(),
		libp2p.NoListenAddrs,
	)
	defer func() {
		_ = client.Close()
	}()

	if err != nil {
		log.Fatal(err)
	}

	client.Connect(context.Background(), serverInfo)

	p := ping.Ping(context.Background(), client, serverInfo.ID)

	pingSoFar := 0
	for pingSoFar < pings {
		res := <-p
		pingSoFar++
		if res.Error != nil {
			log.Fatal(res.Error)
		}
		time.Sleep(5 * time.Second)
	}
}
