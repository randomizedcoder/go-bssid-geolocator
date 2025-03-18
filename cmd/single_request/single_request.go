package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"
)

const (
	bssidCst = "00:11:22:33:44:55"
)

var (
	headers = map[string]string{
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept":          "*/*",
		"Accept-Charset":  "utf-8",
		"Accept-Encoding": "gzip, deflate",
		"Accept-Language": "en-us",
		"User-Agent":      "locationd/1753.17 CFNetwork/711.1.12 Darwin/14.0.0",
	}
)

func main() {

	bssid := flag.String("bssid", "bssidCst", "bssid")

	flag.Parse()

	geoLocate(*bssid)

}

func geoLocate(bssid string) {

	dataBSSID := []byte{0x12, 0x13, 0x0A, 0x11}
	dataBSSID = append(dataBSSID, []byte(bssid)...)
	dataBSSID = append(dataBSSID, []byte{0x18, 0x00, 0x20, 0x00}...)

	data := []byte("\x00\x01\x00\x05en_US\x00\x13com.apple.locationd\x00\x0a" +
		"8.1.12B411\x00\x00\x00\x01\x00\x00\x00")

	data = append(data, byte(len(dataBSSID)))

	data = append(data, dataBSSID...)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://gs-loc.apple.com/clls/wloc", bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Response Status: %s\n", resp.Status)
}
