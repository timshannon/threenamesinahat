// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"net/http"
	"strings"

	"github.com/timshannon/threenamesinahat/game"
	"golang.org/x/net/websocket"
)

func gameTemplate(w *templateWriter, r *http.Request) {
	s := strings.Split(r.URL.Path, "/")
	code := s[len(s)-1]
	// handle trailing slash
	if code == "" {
		code = s[len(s)-2]
	}

	g, ok := game.Find(code)

	if !ok {
		notFound(w, r)
		return
	}
	w.execute(g)
}

func gameSocket(ws *websocket.Conn) {
	websocket.JSON.Send(ws, map[string]interface{}{
		"test": "value",
	})
}
