// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"strings"
	"sync"
	"time"
)

// Player keeps track of a given player as well as is the communication channel
type Player struct {
	sync.Mutex

	Name  string   `json:"name"`
	Names []string `json:"names"`

	chanPing chan bool

	send MsgFunc
}

func (p *Player) Recieve(m Msg) {
	switch strings.ToLower(m.Type) {
	case "pong":
		p.chanPing <- true
	}
}

func (p *Player) update(g *Game) error {
	return p.send(Msg{
		Type: "state",
		Data: g,
	})
}

func (p *Player) ping() bool {
	p.send(Msg{
		Type: "ping",
	})

	select {
	case <-time.After(3 * time.Second):
		return false
	case <-p.chanPing:
		return true
	}
}
