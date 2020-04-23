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
	m := &game.Msg{}
	err := websocket.JSON.Receive(ws, m)
	if err != nil {
		websocket.JSON.Send(ws, &game.Msg{Type: "error", Data: err.Error()})
	}

	if strings.ToLower(m.Type) != "join" {
		ws.Close()
		return
	}

	data, ok := m.Data.(map[string]interface{})
	if !ok {
		websocket.JSON.Send(ws, &game.Msg{Type: "error", Data: "Invalid websocket data"})
		ws.Close()
		return
	}

	player, err := game.Join(data["code"].(string), data["name"].(string), func(m game.Msg) error {
		return websocket.JSON.Send(ws, m)
	})

	if err != nil {
		websocket.JSON.Send(ws, &game.Msg{Type: "error", Data: err.Error()})
		ws.Close()
		return
	}

	for {
		websocket.JSON.Receive(ws, m)
		player.Recieve(*m)
	}

}
