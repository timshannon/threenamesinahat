// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
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
	defer func() {
		g.updatePlayers()
		g.Unlock()
	}()

	if player, ok := g.Team1.player(name); ok {
		if player.ping() {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		player.send = send
		return player, nil
	}

	if player, ok := g.Team2.player(name); ok {
		if player.ping() {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		player.send = send
		return player, nil
	}

	// new player
	if len(g.Team1.Players) <= len(g.Team2.Players) {
		player := g.Team1.addPlayer(name, send)
		if len(g.Team1.Players) == 1 {
			// first player in is leader
			g.Leader = player
		}

		return player, nil
	}

	player := g.Team2.addPlayer(name, send)
	return player, nil
}

func (g *Game) updatePlayers() {
	g.Team1.updatePlayers(g)
	g.Team2.updatePlayers(g)
}

// cleanAndUpdatePlayers pings all players and removes those that don't respond
func (g *Game) cleanAndUpdatePlayers() {
	g.Team1.cleanPlayers()
	g.Team2.cleanPlayers()
	g.updatePlayers()
}

func (g *Game) removePlayer(name string) {
	if !g.Team1.removePlayer(name) {
		g.Team2.removePlayer(name)
	}
	g.updatePlayers()
}
