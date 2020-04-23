// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"net/http"

	"github.com/timshannon/threenamesinahat/game"
	"golang.org/x/net/websocket"
)

var notFound = gzipHandler(templateHandler(emptyTemplate, "notfound.template.html"))
var errorPage = gzipHandler(templateHandler(emptyTemplate, "error.template.html"))

func setupRoutes() {
	get("/css/", serveStatic("css", true))
	get("/js/", serveStatic("js", true))
	get("/", gzipHandler(templateHandler(emptyTemplate, "index.template.html")))
	get("/new", gzipHandler(func(w http.ResponseWriter, r *http.Request) {
		g := game.New()
		http.Redirect(w, r, "/game/"+g.Code, http.StatusTemporaryRedirect)
	}))
	get("/game/", gzipHandler(templateHandler(gameTemplate, "game.template.html")))
	http.Handle("/game", websocket.Handler(gameSocket))
}

func get(pattern string, handler http.HandlerFunc)    { method("GET", pattern, handler) }
func put(pattern string, handler http.HandlerFunc)    { method("PUT", pattern, handler) }
func post(pattern string, handler http.HandlerFunc)   { method("POST", pattern, handler) }
func delete(pattern string, handler http.HandlerFunc) { method("DELETE", pattern, handler) }

func method(method, pattern string, handler http.HandlerFunc) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			notFound(w, r)
			return
		}

		handler(w, r)
	})
}
