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

	go func() {
		//Capture program shutdown, to make sure everything shuts down nicely
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for sig := range c {
			if sig == os.Interrupt {
				log.Print("Three Names in a Hat server is shutting down")
				err := server.Teardown()
				if err != nil {
					log.Fatalf("Error tearing down web server: %s", err)
				}
				os.Exit(0)
			}
		}
	}()
}

func main() {
	flag.Parse()

	log.Printf("Starting web server on port %s", flagPort)
	err := server.Start(flagPort)
	if err != nil {
		log.Fatalf("Error starting web server: %s", err)
	}
}
