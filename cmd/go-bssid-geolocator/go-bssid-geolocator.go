package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/randomizedcoder/go-bssid-geolocator/pkg/geolocator"
)

const (
	debugLevelCst = 111

	signalChannelSizeCst = 10
	cancelSleepTimeCst   = 5 * time.Second

	promListenCst           = ":9088" // [::1]:9088
	promPathCst             = "/metrics"
	promMaxRequestsInFlight = 10
	promEnableOpenMetrics   = true

	concurrentCst = 1
	ouiCst        = "1C:6A:1B" // ubiquity
	countCst      = 1
)

var (
	// Passed by "go build -ldflags" for the show version
	commit  string
	date    string
	version string

	debugLevel uint
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	complete := make(chan struct{}, signalChannelSizeCst)
	go initSignalHandler(cancel, complete)

	concurrent := flag.Uint("concurrent", concurrentCst, "concurrent requests count")
	oui := flag.String("oui", ouiCst, "oui starting point")
	count := flag.Uint64("count", countCst, "count of bssids")

	promListen := flag.String("promListen", promListenCst, "Prometheus http listening socket")
	promPath := flag.String("promPath", promPathCst, "Prometheus http path")

	v := flag.Bool("v", false, "show version")

	d := flag.Uint("d", debugLevelCst, "debug level")

	flag.Parse()

	// Print version information passed in via ldflags in the Makefile
	if *v {
		log.Printf("go-bssid-geolocator commit:%s\tdate(UTC):%s\tversion:%s", commit, date, version)
		os.Exit(0)
	}

	debugLevel = *d

	go initPromHandler(*promPath, *promListen)
	if debugLevel > 10 {
		log.Println("Prometheus http listener started on:", *promListen, *promPath)
	}

	conf := &geolocator.GeoLocatorConf{
		Concurrent: *concurrent,
		Oui:        *oui,
		Count:      *count,
		DebugLevel: debugLevel,
	}

	geolocator := geolocator.NewGeoLocator(ctx, cancel, conf)

	var wg sync.WaitGroup
	wg.Add(1)
	geolocator.Run(&wg)

	wg.Wait()
	complete <- struct{}{}

	if debugLevel > 10 {
		log.Println("xtcp2.go Main complete - farewell")
	}

}

// initSignalHandler sets up signal handling for the process, and
// will call cancel() when received
func initSignalHandler(cancel context.CancelFunc, complete <-chan struct{}) {

	c := make(chan os.Signal, signalChannelSizeCst)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Printf("Signal caught, closing application")
	cancel()

	log.Printf("Signal caught, cancel() called, and sleeping to allow goroutines to close, sleeping:%s",
		cancelSleepTimeCst.String())
	timer := time.NewTimer(cancelSleepTimeCst)

	select {
	case <-complete:
		log.Printf("<-complete exit(0)")
	case <-timer.C:
		// if we exit here, this means all the other go routines didn't shutdown
		// need to investigate why
		log.Printf("Sleep complete, goodbye! exit(0)")
	}

	os.Exit(0)
}

// initPromHandler starts the prom handler with error checking
// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp?tab=doc#HandlerOpts
func initPromHandler(promPath string, promListen string) {
	http.Handle(promPath, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics:   promEnableOpenMetrics,
			MaxRequestsInFlight: promMaxRequestsInFlight,
		},
	))
	go func() {
		err := http.ListenAndServe(promListen, nil)
		if err != nil {
			log.Fatal("prometheus error", err)
		}
	}()
}
