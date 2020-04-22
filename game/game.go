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
	state GameState
}

type GameState struct {
	Code           string  `json:"code"`
	NamesPerPlayer int     `json:"namesPerPlayer"`
	Team1          Team    `json:"team1"`
	Team2          Team    `json:"team2"`
	Leader         *Player `json:"leader"`
}

func (g *Game) Code() string {
	return g.state.Code
}

// join adds a new player to the game
func (g *Game) join(name string, send MsgFunc) (*Player, error) {
	if name == "" {
		return nil, fail.New("You must provide a name before joining")
	}
	g.Lock()
	defer g.Unlock()

	if player, ok := g.state.Team1.player(name); ok {
		if player.connected {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		player.send = send
		return player, nil
	}

	if player, ok := g.state.Team2.player(name); ok {
		if player.connected {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		player.send = send
		return player, nil
	}

	// new player
	if len(g.state.Team1.Players) <= len(g.state.Team2.Players) {
		player := g.state.Team1.addPlayer(name, g, send)
		if len(g.state.Team1.Players) == 1 {
			// first player in is leader
			g.state.Leader = player
		}

		err := g.updatePlayers()
		if err != nil {
			return nil, err
		}

		return player, nil
	}

	player := g.state.Team2.addPlayer(name, g, send)
	err := g.updatePlayers()
	if err != nil {
		return nil, err
	}
	return player, nil
}

func (g *Game) updatePlayers() error {
	err := g.state.Team1.updatePlayers(g)
	if err != nil {
		return err
	}
	err = g.state.Team2.updatePlayers(g)
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
