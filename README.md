Go version of a bssid geolocator, designed for scanning OUI Ranges



## Go code

Code comes mostly from here:

https://github.com/acheong08/apple-corelocation-experiments

The go code has a more complete protobuf:

https://github.com/acheong08/apple-corelocation-experiments/blob/main/pb/BSSIDApple.proto

## Python
There is a python version, that's not as nice

Pythons version:
https://github.com/gigaryte/bssid-geolocator/

Less complete proto:
https://github.com/randomizedcoder/bssid-geolocator/blob/main/bssid.proto

## YouTube
Youtube video is here:

Surveilling the Masses with Wi-Fi Positioning Systems
https://youtu.be/hlbjUvkoyBA?si=uW1vOIopXt-NI98G

## MAC Addressing
MAC addressses

https://en.wikipedia.org/wiki/MAC_address#/media/File:MAC-48_Address.svg

![MAC-48_Address.svg](./doc/MAC-48_Address.svg "mac")


## OUI

Organizationally unique identifier

https://en.wikipedia.org/wiki/Organizationally_unique_identifier


https://standards-oui.ieee.org/oui/oui.txt

```
1C-6A-1B   (hex)		Ubiquiti Inc
1C6A1B     (base 16)		Ubiquiti Inc
				685 Third Avenue, 27th Floor
				New York  NY  New York NY 10017
				US
```

## go OUI

OUI library:
https://github.com/gptlang/oui

## Find BSSID you're connected too?

```
iw dev wlp0s20f3 link
```

```
iwlist wlp0s20f3 scan | grep -A 10 "ESSID:\"MyWiFi\"" | grep "Address"
```

e.g.
```
[das@t:~]$ iw dev wlp0s20f3 link
Connected to 9c:05:d6:<snip> (on wlp0s20f3)
        SSID: <snip>
        freq: 5240.0
        RX: 2112174832 bytes (44635273 packets)
        TX: 2831794242 bytes (16819438 packets)
        signal: -32 dBm
        rx bitrate: 1200.9 MBit/s 80MHz HE-MCS 11 HE-NSS 2 HE-GI 0 HE-DCM 0
        tx bitrate: 1200.9 MBit/s 80MHz HE-MCS 11 HE-NSS 2 HE-GI 0 HE-DCM 0
        bss flags: short-slot-time
        dtim period: 3
        beacon int: 100
```