// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/timshannon/threenamesinahat/game"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
}

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

func gameSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	m := &game.Msg{}
	err = websocket.ReadJSON(ws, m)
	if err != nil {
		websocket.WriteJSON(ws, &game.Msg{Type: "error", Data: err.Error()})
	}

	if strings.ToLower(m.Type) != "join" {
		ws.Close()
		return
	}

	data, ok := m.Data.(map[string]interface{})
	if !ok {
		websocket.WriteJSON(ws, &game.Msg{Type: "error", Data: "Invalid websocket data"})
		ws.Close()
		return
	}

	gameCode := data["code"].(string)
	playerName := data["name"].(string)

	player, err := game.Join(gameCode, playerName)

	if err != nil {
		websocket.WriteJSON(ws, &game.Msg{Type: "error", Data: err.Error()})
		player.Remove()
		ws.Close()
		return
	}

	go func() {
		for msg := range player.Send {
			err := websocket.WriteJSON(ws, msg)
			if err != nil {
				log.Printf("Error in game %s sending to player %s: %s", gameCode, playerName, err)
				player.Remove()
				ws.Close()
				return
			}
		}
	}()

	for {
		err = websocket.ReadJSON(ws, m)
		if err != nil {
			log.Printf("Error recieving on web socket: %s", err)
			player.Remove()
			ws.Close()
			return
		}

		player.Receive <- *m
	}
}
