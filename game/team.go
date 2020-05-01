// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

type Team struct {
	// Name    string    `json:"name"`
	// Color   string    `json:"color"`
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

func (t *Team) addNewPlayer(name string, game *Game) *Player {
	t.Players = append(t.Players, newPlayer(name, game))
	return t.Players[len(t.Players)-1]
}

func (t *Team) addExistingPlayer(player *Player) {
	t.Players = append(t.Players, player)
}

func (t *Team) removePlayer(name string) bool {
	for i := range t.Players {
		if t.Players[i].Name == name {
			t.Players = append(t.Players[:i], t.Players[i+1:]...)
			return true
		}
	}
	return false
}

func (t *Team) updatePlayers() {
	for i := range t.Players {
		t.Players[i].update()
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
