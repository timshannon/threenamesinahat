// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

type Team struct {
	Name    string   `json:"name"`
	Color   string   `json:"color"`
	Players []Player `json:"players"`
}

func (t *Team) player(name string) (*Player, bool) {
	for i := range t.Players {
		if t.Players[i].Name == name {
			return &t.Players[i], true
		}
	}
	return nil, false
}

func (t *Team) addPlayer(name string, game *Game, send MsgFunc) *Player {
	t.Players = append(t.Players, Player{Name: name, send: send, game: game})
	return &t.Players[len(t.Players)-1]
}

func (t *Team) updatePlayers(g *Game) error {
	for _, p := range t.Players {
		err := p.Update()
		if err != nil {
			// TODO: only return error to failed player call? and Log error, instead of stopping whole team update?
			return err
		}
	}
	return nil
}
