// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

// Player keeps track of a given player as well as is the communication channel
type Player struct {
	Name  string   `json:"name"`
	Names []string `json:"names"`
	game  *Game

	connected bool // FIXME?

	send MsgFunc
}

func (p *Player) Update() error {
	return p.send(Msg{
		Type: "state",
		Data: p.game.state,
	})
}
