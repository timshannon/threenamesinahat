// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"strings"
	"sync"
)

// Player keeps track of a given player as well as is the communication channel
type Player struct {
	sync.Mutex

	Name  string   `json:"name"`
	Names []string `json:"names"`
	game  *Game

	connected bool

	send MsgFunc
}

func (p *Player) Recieve(m Msg) {
	switch strings.ToLower(m.Type) {
	case "pong":
		p.Lock()
		p.connected = true
		p.Unlock()
	}
}

func (p *Player) update() error {
	return p.send(Msg{
		Type: "state",
		Data: p.game,
	})
}

func (p *Player) ping() error {
	p.Lock()
	defer p.Unlock()
	p.connected = false
	// TODO wait for pong
	return p.send(Msg{
		Type: "ping",
	})
}
