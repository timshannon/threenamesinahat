// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"log"
	"sync"
)

type Team struct {
	sync.Mutex
	Name    string    `json:"name"`
	Color   string    `json:"color"`
	Players []*Player `json:"players"`
}

func (t *Team) player(name string) (*Player, bool) {
	for i := range t.Players {
		if t.Players[i].Name == name {
			return t.Players[i], true
		}
	}
	return nil, false
}

func (t *Team) addNewPlayer(name string, send MsgFunc, game *Game) *Player {
	t.Lock()
	defer t.Unlock()
	t.Players = append(t.Players, &Player{Name: name, send: send, game: game})
	return t.Players[len(t.Players)-1]
}

func (t *Team) addExistingPlayer(player *Player) {
	t.Lock()
	defer t.Unlock()
	t.Players = append(t.Players, player)
}

func (t *Team) removePlayer(name string) bool {
	t.Lock()
	defer t.Unlock()

	for i := range t.Players {
		if t.Players[i].Name == name {
			t.Players = append(t.Players[:i], t.Players[i+1:]...)
			return true
		}
	}
	return false
}

func (t *Team) updatePlayers(g *Game) {
	for i := range t.Players {
		err := t.Players[i].update()
		if err != nil {
			log.Printf("Error updating player %s: %s", t.Players[i].Name, err)
		}
	}
}

func (t *Team) cleanPlayers() {
	var remove []string
	for i := range t.Players {
		if !t.Players[i].ping() {
			remove = append(remove, t.Players[i].Name)
		}
	}

	for _, name := range remove {
		t.removePlayer(name)
	}
}
