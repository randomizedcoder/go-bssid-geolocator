package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	bssidCst = "1C:6A:1B:00:00:00"
)

// Function to construct the binary request payload
func buildRequestPayload(bssid string) []byte {
	// Convert BSSID from "XX:XX:XX:XX:XX:XX" to raw bytes
	var bssidBytes [6]byte
	fmt.Sscanf(bssid, "%02X:%02X:%02X:%02X:%02X:%02X",
		&bssidBytes[0], &bssidBytes[1], &bssidBytes[2],
		&bssidBytes[3], &bssidBytes[4], &bssidBytes[5])

	// Construct request payload (binary format)
	dataBSSID := append([]byte{0x12, 0x13, 0x0A, 0x11}, bssidBytes[:]...)
	dataBSSID = append(dataBSSID, []byte{0x18, 0x00, 0x20, 0x00}...)

	data := []byte("\x00\x01\x00\x05en_US\x00\x13com.apple.locationd\x00\x0a" +
		"8.1.12B411\x00\x00\x00\x01\x00\x00\x00")

	data = append(data, byte(len(dataBSSID))) // Append length of BSSID data
	data = append(data, dataBSSID...)         // Append BSSID data

	return data
}

// Function to query Apple's Wi-Fi location API
func queryAppleLocation(bssid string) {
	// Build binary request payload
	payload := buildRequestPayload(bssid)

	// Set request headers
	headers := map[string]string{
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept":          "*/*",
		"Accept-Charset":  "utf-8",
		"Accept-Encoding": "gzip, deflate",
		"Accept-Language": "en-us",
		"User-Agent":      "locationd/1753.17 CFNetwork/711.1.12 Darwin/14.0.0",
	}

	// Make HTTP request
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://gs-loc.apple.com/clls/wloc", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check if response is gzip encoded
	var reader io.Reader
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			log.Fatalf("Failed to create gzip reader: %v", err)
		}
	} else {
		reader = resp.Body
	}

	// Read response body
	body, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// Decode binary response to extract latitude & longitude
	parseAppleResponse(body)
}

// Function to parse Apple's binary response
func parseAppleResponse(response []byte) {
	if len(response) < 24 {
		log.Println("Invalid response length")
		return
	}

	// Extract latitude & longitude from binary response
	lat := int64(binary.BigEndian.Uint32(response[10:14]))
	lng := int64(binary.BigEndian.Uint32(response[14:18]))

	// Convert to floating point values
	latitude := float64(lat) * 1e-8
	longitude := float64(lng) * 1e-8

	fmt.Printf("Latitude: %.6f, Longitude: %.6f\n", latitude, longitude)
}

func main() {

	bssid := flag.String("bssid", bssidCst, "oui starting point")

	flag.Parse()

	// Example BSSID (replace with a real one)
	//bssid := "00:1A:2B:3C:4D:5E"

	// Query Apple for location of the BSSID
	queryAppleLocation(*bssid)
}
