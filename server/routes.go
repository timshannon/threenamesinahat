package server

import (
	"net/http"
)

func setupRoutes() {
	http.HandleFunc("/css/", serveStatic("css", true))
	http.HandleFunc("/js/", serveStatic("js", true))
}
