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

const (
	stagePregame = "pregame" // players join
	stageSetup   = "setup"   // players add names
	stageRound1  = "round1"
	stageRound2  = "round2"
	stageRound3  = "round3"
)

type Game struct {
	sync.Mutex

	Code           string  `json:"code"`
	NamesPerPlayer int     `json:"namesPerPlayer"`
	Team1          Team    `json:"team1"`
	Team2          Team    `json:"team2"`
	Leader         *Player `json:"leader"`
	Stage          string  `json:"stage"`
}

// pregame -> setup

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
	if g.Stage != stagePregame {
		return nil, fail.New("You cannot join a game in progress")
	}

	if len(g.Team1.Players) <= len(g.Team2.Players) {
		player := g.Team1.addNewPlayer(name, send, g)
		if len(g.Team1.Players) == 1 {
			// first player in is leader
			g.Leader = player
		}

		return player, nil
	}

	player := g.Team2.addNewPlayer(name, send, g)
	return player, nil
}

func (g *Game) setNamesPerPlayer(who *Player, num int) error {
	if g.Stage != stagePregame {
		return fail.New("The number of names per player cannot be set after the game has started")
	}

	if !who.isLeader() {
		return fail.New("Only game leaders can change the number of names")
	}

	if num <= 0 {
		return fail.New("Number of names must be greater than 0")
	}

	if num > 20 {
		return fail.New("The maximum number of names is 20")
	}

	g.Lock()
	defer func() {
		g.updatePlayers()
		g.Unlock()
	}()

	g.NamesPerPlayer = num
	return nil
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

func (g *Game) startGame(who *Player) error {
	if !who.isLeader() {
		return fail.New("Only the %s can start the game", g.Leader.Name)
	}
	if g.Stage != stagePregame {
		return fail.New("The game has already started")
	}

	g.cleanAndUpdatePlayers()

	if len(g.Team1.Players) < 2 || len(g.Team2.Players) < 2 {
		return fail.New("You must have at least two players on each team to start")
	}

	g.Lock()
	defer func() {
		g.updatePlayers()
		g.Unlock()
	}()
	g.Stage = stageSetup

	return nil
}

func (g *Game) switchTeams(who *Player) {
	g.Lock()
	defer func() {
		g.updatePlayers()
		g.Unlock()
	}()

	if g.Team1.removePlayer(who.Name) {
		g.Team2.addExistingPlayer(who)
	} else {
		g.Team2.removePlayer(who.Name)
		g.Team1.addExistingPlayer(who)
	}
}
