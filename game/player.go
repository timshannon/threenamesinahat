// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"log"
	"strings"
	"time"

	"github.com/timshannon/threenamesinahat/fail"
)

const timeout = 3 * time.Second

// Player keeps track of a given player as well as is the communication channel
type Player struct {
	Name  string   `json:"name"`
	Names []string `json:"names"`

	chanPing chan bool

	Send    chan Msg `json:"-"`
	Receive chan Msg `json:"-"`

	game *Game
}

func newPlayer(name string, game *Game) *Player {
	p := &Player{
		Name:     name,
		Send:     make(chan Msg, 2),
		Receive:  make(chan Msg, 2),
		chanPing: make(chan bool),
		game:     game,
	}

	go recieve(p)

	return p

}

func recieve(p *Player) {
	for msg := range p.Receive {
		go func(m Msg) {
			switch strings.ToLower(m.Type) {
			case "pong":
				p.chanPing <- true
			case "namesperplayer":
				if _, ok := m.Data.(float64); !ok {
					p.ok(fail.New("Invalid data type for namesperplayer. Got %T wanted float64", m.Data))
					break
				}
				p.ok(p.game.setNamesPerPlayer(p, int(m.Data.(float64))))
			case "start":
				p.ok(p.game.startGame(p))
			case "switchteams":
				p.game.switchTeams(p)
			default:
				p.ok(fail.New("%s is an invalid message type", m.Type))
			}
		}(msg)
	}
}

func (p *Player) ok(err error) bool {
	if err != nil {
		if fail.IsFailure(err) {
			p.Send <- Msg{
				Type: "error",
				Data: err.Error(),
			}
		} else {
			log.Printf("Internal Error: %s", err)
			p.Send <- Msg{
				Type: "error",
				Data: "An internal error has occured, please start a new game",
			}
		}
		return true
	}
	return false
}

func (p *Player) update() {
	p.Send <- Msg{
		Type: "state",
		Data: p.game,
	}
}

func (p *Player) ping() bool {
	go func() { p.Send <- Msg{Type: "ping"} }()

	select {
	case <-time.After(timeout):
		return false
	case <-p.chanPing:
		return true
	}
}

func (p *Player) isLeader() bool {
	return p.Name == p.game.Leader.Name
}
