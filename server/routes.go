package server

import "net/http"

func setupRoutes() http.Handler {
	return http.NotFoundHandler()
}
