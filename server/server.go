package server

import (
	"compress/gzip"
	"context"
	"net/http"
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

func Start(port string) error {
	server = &http.Server{
		Handler: setupRoutes(),
		Addr:    ":" + port,
	}
	return server.ListenAndServe()
}

func Teardown() error {
	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		// TODO: Teardown Websockets
		return server.Shutdown(ctx)
	}
	return nil
}
