// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/timshannon/threenamesinahat/game"
	"golang.org/x/net/websocket"
)

var gameNotFound = gzipHandler(templateHandler(func(w *templateWriter, r *http.Request) {
	w.execute(struct {
		GameNotFound bool
	}{
		GameNotFound: true,
	})
}, "notfound.template.html"))

func gameTemplate(w *templateWriter, r *http.Request) {
	s := strings.Split(r.URL.Path, "/")
	code := s[len(s)-1]
	// handle trailing slash
	if code == "" {
		code = s[len(s)-2]
	}

	g, ok := game.Find(code)

	if !ok {
		gameNotFound(w, r)
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

	gameCode := data["code"].(string)
	playerName := data["name"].(string)

	player, err := game.Join(gameCode, playerName)

	if err != nil {
		websocket.JSON.Send(ws, &game.Msg{Type: "error", Data: err.Error()})
		ws.Close()
		return
	}

	go func() {
		for msg := range player.Send {
			err := websocket.JSON.Send(ws, msg)
			if err != nil {
				log.Printf("Error in game %s sending to player %s: %s", gameCode, playerName, err)
				ws.Close()
				return
			}
		}
	}()

	for {
		err = websocket.JSON.Receive(ws, m)
		if err != nil {
			log.Printf("Error recieving on web socket: %s", err)
			ws.Close()
			return
		}

		player.Receive <- *m
	}
}
