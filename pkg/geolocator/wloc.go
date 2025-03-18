package geolocator

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	bssidv1 "github.com/randomizedcoder/go-bssid-geolocator/protos/bssid/v1"
	"google.golang.org/protobuf/proto"
)

// "github.com/randomizedcoder/go-bssid-geolocator/bssidv1"

const (
	OperatingSystemCst = "iPhone OS17.5/21F79"
	ModelCst           = "iPhone12,1"
)

var headers = map[string]string{
	"Content-Type":    "application/x-www-form-urlencoded",
	"Accept":          "*/*",
	"Accept-Charset":  "utf-8",
	"Accept-Encoding": "gzip",
	//"Accept-Encoding": "gzip, deflate",
	"Accept-Language": "en-us",
	"User-Agent":      "locationd/2890.16.16 CFNetwork/1496.0.7 Darwin/23.5.0",
	//"Keep-Alive":      "timeout=300, max=100000",
}

func (g *GeoLocator) initBytes() {
	var err error
	if g.initialWlocBytes, err = hex.DecodeString("0001000a656e2d3030315f3030310013636f6d2e6170706c652e6c6f636174696f6e64000c31372e352e312e323146393000000001000000"); err != nil {
		panic(err)
	}
	// if g.initialPbcWlocBytes, err = hex.DecodeString("0001000a656e2d3030315f3030310013636f6d2e6170706c652e6c6f636174696f6e64000d31372e342e312e323145323336000000640000"); err != nil {
	// 	panic(err)
	// }
}

func (g *GeoLocator) BuildAppleWLoc(bssids []string) (block *bssidv1.AppleWLoc) {

	var zero int32

	block = &bssidv1.AppleWLoc{
		NumCellResults: &zero,
		DeviceType: &bssidv1.DeviceType{
			OperatingSystem: OperatingSystemCst,
			Model:           ModelCst,
		},
		NumWifiResults: &zero,
	}

	for _, bssid := range bssids {
		block.WifiDevices = append(block.WifiDevices, &bssidv1.WifiDevice{Bssid: bssid})
	}

	return block
}

func (g *GeoLocator) RequestWloc(block *bssidv1.AppleWLoc) (*bssidv1.AppleWLoc, error) {

	serializedBlock, err := g.SerializeProto(block, g.initialWlocBytes)
	if err != nil {
		return nil, errors.New("failed to serialize protobuf")
	}
	var wlocURL string = "https://gs-loc.apple.com"
	// switch args.region {
	// case Options.China:
	// 	log.Println("Using China API")
	// 	wlocURL = "https://gs-loc-cn.apple.com"
	// }

	wlocURL = wlocURL + "/clls/wloc"

	req, _ := http.NewRequest(http.MethodPost, wlocURL, bytes.NewReader(serializedBlock))

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	if g.debugLevel > 10 {
		log.Println("RequestWloc request built")
	}

	if g.debugLevel > 10 {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Fatalf("Error dumping request: %v", err)
		}
		log.Println("Request Dump:\n", string(dump))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("failed to make request")
	}
	defer resp.Body.Close()

	if g.debugLevel > 10 {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatalf("Error dumping response: %v", err)
		}
		log.Println("Response Dump:\n", string(dump))
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 0 {
			return nil, errors.New("cors issue probably")
		}
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}

	var body []byte

	if resp.Header.Get("Content-Encoding") == "gzip" {
		log.Println("Response is gzip-encoded. Decompressing...")

		// Wrap response body with gzip reader
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gzipReader.Close()

		// Read decompressed response
		body, err := io.ReadAll(gzipReader)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
		}

		log.Println("Decompressed Response:\n", string(body))
	} else {
		// If not gzip-encoded, read normally
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
		}
		log.Println("Response:\n", string(body))
	}

	if g.debugLevel > 10 {
		log.Printf("RequestWloc len(body):%d", len(body))
	}

	respBlock := bssidv1.AppleWLoc{}
	err = proto.Unmarshal(body[10:], &respBlock)
	if err != nil {
		return nil, errors.New("failed to unmarshal response protobuf")
	}

	return &respBlock, nil
}
