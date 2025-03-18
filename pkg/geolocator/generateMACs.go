package geolocator

import (
	"fmt"
	"log"
	"net"
)

func (g *GeoLocator) generateMACs(oui string, ch chan<- string) {

	ouiBytes, err := net.ParseMAC(oui + ":00:00:00")
	if err != nil {
		log.Fatal("Invalid OUI format:", err)
		return
	}

	ouiPrefix := ouiBytes[:3]

	for i := 0; i < int(g.conf.Count) && i <= 0xFFFFFF; i++ {

		g.pC.WithLabelValues("generateMACs", "i", "counter").Inc()

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
