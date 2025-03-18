#
# Makefile
#

.PHONY: protos run

all: run

run:
	go run ./cmd/go-bssid-geolocator/go-bssid-geolocator.go

protos:
	./generate.bash

# end
