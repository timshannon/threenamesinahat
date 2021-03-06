// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"net/http"

	"github.com/timshannon/threenamesinahat/game"
)

var notFound = gzipHandler(templateHandler(emptyTemplate, "notfound.template.html"))
var errorPage = gzipHandler(templateHandler(emptyTemplate, "error.template.html"))

func setupRoutes() {
	get("/css/", serveStatic("css", true))
	get("/js/", serveStatic("js", true))
	get("/audio/", serveStatic("audio", true))
	get("/", gzipHandler(templateHandler(emptyTemplate, "index.template.html")))
	get("/new", gzipHandler(func(w http.ResponseWriter, r *http.Request) {
		g, err := game.New(ipAddress(r))
		if err != nil {
			errorPage(w, r)
			return
		}

		http.Redirect(w, r, "/game/"+g.Code, http.StatusTemporaryRedirect)
	}))
	get("/game/", gzipHandler(templateHandler(gameTemplate, "game.template.html")))
	http.HandleFunc("/game", gameSocket)
	get("/join", gzipHandler(templateHandler(emptyTemplate, "join.template.html")))
	get("/about", gzipHandler(templateHandler(aboutTemplate, "about.template.html")))
	get("/rules", gzipHandler(templateHandler(emptyTemplate, "rules.template.html")))
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
