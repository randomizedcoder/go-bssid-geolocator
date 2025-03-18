package geolocator

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	bssidv1 "github.com/randomizedcoder/go-bssid-geolocator/protos/bssid/v1"
)

func (g *GeoLocator) Worker(wg *sync.WaitGroup, id uint, workQueue <-chan string) {

	defer wg.Done()

	ids := fmt.Sprintf("%d", id)

	startTime := time.Now()
	defer func() {
		g.pwH.WithLabelValues(ids, "Worker", "complete", "count").Observe(time.Since(startTime).Seconds())
	}()
	g.pwC.WithLabelValues(ids, "Worker", "start", "count").Inc()

breakPoint:
	for {
		g.pwC.WithLabelValues(ids, "Worker", "for", "count").Inc()

		select {

		case <-g.ctx.Done():
			break breakPoint

		case bssid := <-workQueue:

			if g.debugLevel > 1000 {
				log.Printf("Worker id:%d bssid:%s", id, bssid)
			}

			var err error
			var repBlock *bssidv1.AppleWLoc
			for attempt := 0; attempt < g.conf.Retries; attempt++ {
				repBlock, err = g.RequestWloc(g.BuildAppleWLoc([]string{bssid}))

				if err == nil {
					if attempt == 0 {
						g.pwC.WithLabelValues(ids, "Worker", "firstShot", "count").Inc()
					}
					break // Success, no need to retry
				}

				g.pwC.WithLabelValues(ids, "Worker", "RequestWloc", "error").Inc()
				log.Printf("Worker id:%d Attempt %d failed for BSSID %s: %v", id, attempt, bssid, err)

				// Sleep before retrying (configurable)
				time.Sleep(g.conf.RetrySleepDuration)

				g.pwC.WithLabelValues(ids, "Worker", "retry", "count").Inc()
			}

			if err != nil {
				g.pwC.WithLabelValues(ids, "Worker", "RequestWloc", "error").Inc()
				//log.Fatalf("Worker id:%d RequestWloc err:%v", id, err)
				g.retryWorkQueue <- bssid
				continue
			}

			if g.debugLevel > 1000 {
				log.Printf("Worker id:%d len(repBlock.GetWifiDevices()):%d", id, len(repBlock.GetWifiDevices()))
			}

			g.pwC.WithLabelValues(ids, "Worker", "GetWifiDevices", "count").Add(float64(len(repBlock.GetWifiDevices())))

			for _, ap := range repBlock.GetWifiDevices() {
				lat := float64(*ap.Location.Latitude) * math.Pow(10, -8)
				long := float64(*ap.Location.Longitude) * math.Pow(10, -8)
				// if g.debugLevel > 10 {
				// 	man, err := ouidb.Lookup(ap.Bssid)
				// 	if err != nil {
				// 		man = "Unknown"
				// 	}
				// 	log.Printf("Worker id:%d BSSID: %s (%s) found at Lat: %f Long: %f\n", id, ap.Bssid, man, lat, long)
				// 	continue
				// }
				// fatal error: concurrent map writes

				// goroutine 20 [running]:
				// github.com/gptlang/oui/ouidb.loadDatabase()
				// 				/home/das/go/pkg/mod/github.com/gptlang/oui@v0.0.0-20240522122259-08e97ad0b56a/ouidb/ouidb.go:45 +0x31b
				// github.com/gptlang/oui/ouidb.Lookup({0xc000147050?, 0xc020000000000000?})
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
