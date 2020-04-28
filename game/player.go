// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/timshannon/threenamesinahat/fail"
)

// Player keeps track of a given player as well as is the communication channel
type Player struct {
	sync.Mutex

	Name  string   `json:"name"`
	Names []string `json:"names"`

	chanPing chan bool
	send     MsgFunc
	game     *Game
}

func (p *Player) Recieve(m Msg) {
	switch strings.ToLower(m.Type) {
	case "pong":
		p.chanPing <- true
	case "namesperplayer":
		if _, ok := m.Data.(float64); !ok {
			p.ok(fail.New("Invalid data type for namesperplayer. Got %T wanted float64", m.Data))
			return
		}
		p.ok(p.game.setNamesPerPlayer(p, int(m.Data.(float64))))
	case "start":
		p.ok(p.game.startGame(p))
	case "switchteams":
		p.game.switchTeams(p)
	default:
		p.ok(fail.New("%s is an invalid message type", m.Type))
	}
}

func (p *Player) ok(err error) bool {
	if err != nil {
		if fail.IsFailure(err) {
			p.send(Msg{
				Type: "error",
				Data: err.Error(),
			})
		} else {
			p.send(Msg{
				Type: "error",
				Data: "An internal error has occured, please start a new game",
			})
			log.Printf("Internal Error: %s", err)
		}
		return true
	}
	return false
}

func (p *Player) update() error {
	return p.send(Msg{
		Type: "state",
		Data: p.game,
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

func (p *Player) isLeader() bool {
	return p.Name == p.game.Leader.Name
}
