// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"compress/gzip"
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	server  *http.Server
	zipPool sync.Pool
)

func init() {
	zipPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}
}

func Start(port string, shutdown chan bool) error {
	setupRoutes()
	server = &http.Server{
		Addr: ":" + port,
	}

	err := make(chan error)
	go func() {
		err <- server.ListenAndServe()
	}()

	go func() {
		<-shutdown
		err <- teardown()
	}()

	return <-err
}

func teardown() error {
	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		// TODO: Teardown Websockets
		err := server.Shutdown(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// gzipResponse gzips the response data for any respones writers defined to use it
type gzipResponse struct {
	zip *gzip.Writer
	http.ResponseWriter
}

func (g *gzipResponse) Write(b []byte) (int, error) {
	if g.zip == nil {
		return g.ResponseWriter.Write(b)
	}
	return g.zip.Write(b)
}

func (g *gzipResponse) Close() error {
	if g.zip == nil {
		return nil
	}
	err := g.zip.Close()
	if err != nil {
		return err
	}
	zipPool.Put(g.zip)
	return nil
}

func gzipResponseWriter(w http.ResponseWriter, r *http.Request) *gzipResponse {
	var writer *gzip.Writer

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := zipPool.Get().(*gzip.Writer)
		gz.Reset(w)

		writer = gz
	}

	return &gzipResponse{zip: writer, ResponseWriter: w}
}

func gzipHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if w.Header().Get("Content-Encoding") != "gzip" {
			// only create gzip writer if one doesn't already exist in the handler heirarchy
			w = gzipResponseWriter(w, r)
			defer func() {
				err := w.(*gzipResponse).Close()
				if err != nil {
					log.Printf("Error closing gzip responseWriter: %s", err)
				}
			}()
		}
		handler(w, r)
	}
}
