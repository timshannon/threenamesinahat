package server

import (
	"net/http"
)

func setupRoutes() {
	get("/css/", serveStatic("css", true))
	get("/js/", serveStatic("js", true))
	get("/", gzipHandler(templateHandler(index, "index.template.html")))
}

func get(pattern string, handler http.HandlerFunc)    { method("GET", pattern, handler) }
func put(pattern string, handler http.HandlerFunc)    { method("PUT", pattern, handler) }
func post(pattern string, handler http.HandlerFunc)   { method("POST", pattern, handler) }
func delete(pattern string, handler http.HandlerFunc) { method("DELETE", pattern, handler) }

func method(method, pattern string, handler http.HandlerFunc) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.NotFound(w, r)
			return
		}

		handler(w, r)
	})
}
