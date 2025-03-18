package geolocator

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/gptlang/oui/ouidb"
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

	workQueue chan string

	debugLevel uint
}

type GeoLocatorConf struct {
	Concurrent uint
	Oui        string
	Count      uint64

	DebugLevel uint
}

func NewGeoLocator(ctx context.Context, cancel context.CancelFunc, conf *GeoLocatorConf) *GeoLocator {

	g := new(GeoLocator)

	g.ctx = ctx
	g.cancel = cancel

	g.conf = conf

	g.debugLevel = conf.DebugLevel

	g.initBytes()

	g.workQueue = make(chan string, g.conf.Concurrent)

	return g
}

func (g *GeoLocator) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	var w sync.WaitGroup

	w.Add(int(g.conf.Concurrent))
	for i := uint(0); i < g.conf.Concurrent; i++ {
		go g.Worker(&w, i)
	}

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

func (g *GeoLocator) generateMACs(oui string, ch chan<- string) {

	ouiBytes, err := net.ParseMAC(oui + ":00:00:00")
	if err != nil {
		log.Fatal("Invalid OUI format:", err)
		return
	}

	ouiPrefix := ouiBytes[:3]

	for i := 0; i < int(g.conf.Concurrent) && i <= 0xFFFFFF; i++ {

		// Generate the last 3 bytes
		mac := append(ouiPrefix, byte(i>>16), byte(i>>8), byte(i))

		macStr := fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
			mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])

		if g.debugLevel > 10 {
			log.Printf("generateMACs i:%d macStr:%s", i, macStr)
			ch <- macStr
		}
	}
}

func (g *GeoLocator) Worker(wg *sync.WaitGroup, id uint) {

	defer wg.Done()

breakPoint:
	for {
		select {

		case <-g.ctx.Done():
			break breakPoint

		case bssid := <-g.workQueue:

			if g.debugLevel > 10 {
				log.Printf("Worker id:%d bssid:%s", id, bssid)
			}

			// please note we are only using a single bssid at this point
			repBlock, err := g.RequestWloc(g.BuildAppleWLoc([]string{bssid}))
			if err != nil {
				log.Fatalf("Worker id:%d RequestWloc err:%v", id, err)
			}
			if g.debugLevel > 10 {
				log.Printf("Worker id:%d repBlock", id)
			}
			for _, ap := range repBlock.GetWifiDevices() {
				lat := float64(*ap.Location.Latitude) * math.Pow(10, -8)
				long := float64(*ap.Location.Longitude) * math.Pow(10, -8)
				if g.debugLevel > 10 {
					man, err := ouidb.Lookup(ap.Bssid)
					if err != nil {
						man = "Unknown"
					}
					log.Printf("Worker id:%d BSSID: %s (%s) found at Lat: %f Long: %f\n", id, ap.Bssid, man, lat, long)
					continue
				}
				log.Printf("Worker id:%d BSSID: %s found at Lat: %f Long: %f\n", id, ap.Bssid, lat, long)
			}
		}

		// default:
		// blocking
	}

	log.Printf("worker id:%d done", id)
}

func (g *GeoLocator) FakeWorker(wg *sync.WaitGroup, id uint) {

	defer wg.Done()

	randomDuration := time.Duration(rand.Intn(500)+10) * time.Millisecond
	time.Sleep(sleepDurationCst + randomDuration)

	log.Printf("worker id:%d done", id)

}
