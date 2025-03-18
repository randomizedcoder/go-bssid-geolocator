package geolocator

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// 	"github.com/randomizedcoder/go-bssid-geolocator/bssidv1"
//	"github.com/gptlang/oui/ouidb"

const (
	sleepDurationCst = 5 * time.Second
)

type GeoLocator struct {
	ctx    context.Context
	cancel context.CancelFunc

	conf *GeoLocatorConf

	initialWlocBytes []byte
	//initialPbcWlocBytes []byte

	workQueue      chan string
	retryWorkQueue chan string

	pC *prometheus.CounterVec
	pH *prometheus.SummaryVec
	pG prometheus.Gauge

	pwC *prometheus.CounterVec
	pwH *prometheus.SummaryVec
	pwG prometheus.Gauge

	debugLevel uint
}

type GeoLocatorConf struct {
	Concurrent uint
	Oui        string
	Count      uint64

	Retries            int
	RetrySleepDuration time.Duration

	DebugLevel uint
}

func NewGeoLocator(ctx context.Context, cancel context.CancelFunc, conf *GeoLocatorConf) *GeoLocator {

	g := new(GeoLocator)

	g.ctx = ctx
	g.cancel = cancel

	g.conf = conf

	g.debugLevel = conf.DebugLevel

	g.initBytes()
	g.InitPromethus()

	g.workQueue = make(chan string, g.conf.Concurrent)
	g.retryWorkQueue = make(chan string, g.conf.Concurrent)

	return g
}

func (g *GeoLocator) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	var w sync.WaitGroup

	w.Add(int(g.conf.Concurrent))
	for i := uint(0); i < g.conf.Concurrent; i++ {
		go g.Worker(&w, i, g.workQueue)
	}

	w.Add(1)
	go g.Worker(&w, 666, g.retryWorkQueue)

	if g.debugLevel > 10 {
		log.Println("Run workers started")
	}

	w.Add(1)
	go g.generateMACs(g.conf.Oui, g.workQueue)

	if g.debugLevel > 10 {
		log.Println("Run generateMACs started")
	}

	w.Wait()

}
