// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/timshannon/threenamesinahat/fail"
)

const clientTimeout = 3 * time.Second

// Player keeps track of a given player as well as is the communication channel
type Player struct {
	sync.RWMutex
	playerState

	chanPing chan bool

	Send    chan Msg `json:"-"`
	Receive chan Msg `json:"-"`

	game *Game
}

type playerState struct {
	Name  string   `json:"name"`
	Names []string `json:"names"`
}

func newPlayer(name string, game *Game) *Player {
	p := &Player{
		playerState: playerState{
			Name: name,
		},
		Send:     make(chan Msg, 5),
		Receive:  make(chan Msg, 5),
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
				if num, ok := m.Data.(float64); ok {
					p.ok(p.game.setNamesPerPlayer(p, int(num)))
				} else {
					p.ok(fail.New("Invalid data type for namesperplayer. Got %T wanted float64", m.Data))
				}
			case "start":
				p.ok(p.game.startGame(p))
			case "switchteams":
				p.game.switchTeams(p)
			case "addname":
				if name, ok := m.Data.(string); ok {
					p.ok(p.addName(name))
				} else {
					p.ok(fail.New("Invalid data type for addname.  Got %T wanted string", m.Data))
				}
			case "removename":
				if name, ok := m.Data.(string); ok {
					p.ok(p.removeName(name))
				} else {
					p.ok(fail.New("Invalid data type for removename.  Got %T wanted string", m.Data))
				}
			case "startturn":
				p.ok(p.game.startTurn(p))
			case "nextname":
				p.ok(p.game.nextName(p))
			case "stealyes":
				p.ok(p.game.stealConfirm(p, true))
			case "stealno":
				p.ok(p.game.stealConfirm(p, false))
			case "reset":
				p.ok(p.game.reset(p, ""))
			case "requestupdate":
				p.update(p.game.gameState)
			default:
				p.ok(fail.New("%s is an invalid message type", m.Type))
			}
		}(msg)
	}
}

func (p *Player) ok(err error) bool {
	if err != nil {
		if fail.IsFailure(err) {
			p.SendMsg(Msg{
				Type: "error",
				Data: err.Error(),
			})
		} else {
			log.Printf("Internal Error: %s", err)
			p.SendMsg(Msg{
				Type: "error",
				Data: "An internal error has occured, please start a new game",
			})
		}
		return true
	}
	return false
}

func (p *Player) update(state gameState) {
	p.SendMsg(Msg{
		Type: "state",
		Data: state,
	})
}

func (p *Player) ping() bool {
	go func() { p.SendMsg(Msg{Type: "ping"}) }()

	select {
	case <-time.After(clientTimeout):
		return false
	case <-p.chanPing:
		return true
	}
}

func (p *Player) isLeader() bool {
	p.RLock()
	defer p.RUnlock()
	return p.Name == p.game.Leader.Name
}

func (p *Player) names() []nameItem {
	p.RLock()
	defer p.RUnlock()
	names := make([]nameItem, 0, len(p.Name))

	for i := range p.Names {
		names = append(names, nameItem{name: p.Names[i], player: p.Name})
	}

	return names
}

func (p *Player) addName(name string) error {
	p.Lock()
	defer p.Unlock()

	if p.game.Stage != stageSetup {
		return fail.New("You cannot add names at this time")
	}

	if len(p.Names) >= p.game.NamesPerPlayer {
		return fail.New("You cannot add any more names in this game")
	}

	p.Names = append(p.Names, name)
	p.game.updatePlayers()
	return nil
}

func (p *Player) removeName(name string) error {
	p.Lock()
	defer p.Unlock()
	if p.game.Stage != stageSetup {
		return fail.New("You cannot remove names at this time")
	}

	for i := range p.Names {
		if p.Names[i] == name {
			p.Names = append(p.Names[:i], p.Names[i+1:]...)
			p.game.updatePlayers()
			return nil
		}
	}

	return nil
}

func (p *Player) clearNames() {
	p.Lock()
	defer p.Unlock()
	p.Names = nil
}

func (p *Player) playSound(sound string) {
	p.SendMsg(Msg{
		Type: "playsound",
		Data: sound,
	})
}

func (p *Player) sendNotification(notification string) {
	p.SendMsg(Msg{
		Type: "notification",
		Data: notification,
	})
}

// SendMsg sends a Msg to a player
func (p *Player) SendMsg(msg Msg) {
	go func() {
		if p == nil {
			return
		}
		p.Send <- msg
	}()
}

// MarshalJSON implements the JSON Marshaller interface to allow managing locks when the data is marshalled
func (p *Player) MarshalJSON() ([]byte, error) {
	p.RLock()
	defer p.RUnlock()
	return json.Marshal(p.playerState)
}
