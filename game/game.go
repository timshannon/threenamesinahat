// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"math/rand"
	"sync"

	"github.com/timshannon/threenamesinahat/fail"
)

const codeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Msg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type MsgFunc func(Msg) error

type Game struct {
	sync.Mutex

	Code           string  `json:"code"`
	NamesPerPlayer int     `json:"namesPerPlayer"`
	Team1          Team    `json:"team1"`
	Team2          Team    `json:"team2"`
	Leader         *Player `json:"leader"`
}

// join adds a new player to the game
func (g *Game) join(name string, send MsgFunc) (*Player, error) {
	if name == "" {
		return nil, fail.New("You must provide a name before joining")
	}
	g.Lock()
	defer g.Unlock()

	if player, ok := g.Team1.player(name); ok {
		if player.connected {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		player.send = send
		return player, nil
	}

	if player, ok := g.Team2.player(name); ok {
		if player.connected {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		player.send = send
		return player, nil
	}

	// new player
	if len(g.Team1.Players) <= len(g.Team2.Players) {
		player := g.Team1.addPlayer(name, g, send)
		if len(g.Team1.Players) == 1 {
			// first player in is leader
			g.Leader = player
		}

		err := g.updatePlayers()
		if err != nil {
			return nil, err
		}

		return player, nil
	}

	player := g.Team2.addPlayer(name, g, send)
	err := g.updatePlayers()
	if err != nil {
		return nil, err
	}
	return player, nil
}

func (g *Game) updatePlayers() error {
	err := g.Team1.updatePlayers(g)
	if err != nil {
		return err
	}
	err = g.Team2.updatePlayers(g)
	return err
}

// generateCode generates a random string of only uppercase letters (A-Z) of the specified length
func generateCode(length int) string {
	code := ""

	for i := 0; i < length; i++ {
		code += string(codeChars[rand.Intn(26)])
	}

	return code
}
