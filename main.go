//go:generate go run -tags=dev files/assets_generate.go
// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/timshannon/threenamesinahat/server"
)

var flagPort string

func init() {
	flag.StringVar(&flagPort, "port", "8080", "Port for the webserver to listen on")
}

func main() {
	flag.Parse()

	//Capture program shutdown, to make sure everything shuts down nicely
	c := make(chan os.Signal, 1)
	shutdown := make(chan bool)

	go func() {
		signal.Notify(c, os.Interrupt)
		for sig := range c {
			if sig == os.Interrupt {
				log.Print("Three Names in a Hat server is shutting down")
				shutdown <- true
			}
		}
	}()

	log.Printf("Starting web server on port %s", flagPort)
	err := server.Start(flagPort, shutdown)
	if err != nil {
		log.Fatalf("Fatal web server error: %s", err)
	}
	log.Print("Server Shutdown Successfully")
	os.Exit(0)
}
