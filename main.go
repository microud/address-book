package main

import (
	"flag"
	"time"
)

func main() {
	var device string
	flag.StringVar(&device, "device", "eth0", "device for traffic capturing")
	flag.Parse()

	db, err := NewDatabase()
	if err != nil {
		panic(err)
	}

	capture, err := NewCapture(db, CaptureParams{
		Device:         device,
		SnapshotLength: 1024,
		Promiscuous:    false,
		Timeout:        30 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	go capture.Run()

	server := NewServer(db)

	server.Run()
}
