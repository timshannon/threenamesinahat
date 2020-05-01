// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"sync"
	"time"

	"github.com/timshannon/threenamesinahat/fail"
)

type Msg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// pregame -> setup -> round1 -> round2 -> round3 -> end
const (
	stagePregame = "pregame" // players join
	stageSetup   = "setup"   // players add names
	stageRound1  = "round1"
	stageRound2  = "round2"
	stageRound3  = "round3"
	stageEnd     = "end"
)

const (
	timePerTurn      = 60 * time.Second // how much time each player gets per turn
	setupTimePerName = 30 * time.Second // how much time per name each player gets during game setup
)

type Game struct {
	sync.RWMutex
	gameState
}

type gameState struct {
	Code           string  `json:"code"`
	NamesPerPlayer int     `json:"namesPerPlayer"`
	Team1          Team    `json:"team1"`
	Team2          Team    `json:"team2"`
	Leader         *Player `json:"leader"`
	Stage          string  `json:"stage"`
	Timer          struct {
		Seconds int `json:"seconds"`
		Left    int `json:"left"`
	} `json:"timer"`

	nameList []string
}

// join adds a new player to the game
func (g *Game) join(name string) (*Player, error) {
	if name == "" {
		return nil, fail.New("You must provide a name before joining")
	}
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

	if player, ok := g.Team1.player(name); ok {
		if player.ping() {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		return player, nil
	}

	if player, ok := g.Team2.player(name); ok {
		if player.ping() {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		return player, nil
	}

	// new player
	if g.Stage != stagePregame {
		return nil, fail.New("You cannot join a game in progress")
	}

	if len(g.Team1.Players) <= len(g.Team2.Players) {
		player := g.Team1.addNewPlayer(name, g)
		if len(g.Team1.Players) == 1 {
			// first player in is leader
			g.Leader = player
		}

		return player, nil
	}

	player := g.Team2.addNewPlayer(name, g)
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
		g.Unlock()
		g.updatePlayers()
	}()

	g.NamesPerPlayer = num
	return nil
}

func (g *Game) updatePlayers() {
	g.RLock()
	defer g.RUnlock()
	g.Team1.updatePlayers()
	g.Team2.updatePlayers()
}

func (g *Game) removePlayer(name string) {
	// TODO: If player is leader, make a new leader?
	if !g.Team1.removePlayer(name) {
		g.Team2.removePlayer(name)
	}
	g.updatePlayers()
}

func (g *Game) startGame(who *Player) error {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

	if !who.isLeader() {
		return fail.New("Only the %s can start the game", g.Leader.Name)
	}
	if g.Stage != stagePregame {
		return fail.New("The game has already started")
	}

	g.Team1.cleanPlayers()
	g.Team2.cleanPlayers()

	if len(g.Team1.Players) < 2 || len(g.Team2.Players) < 2 {
		// return to pregame and wait for players to join
		return nil
	}

	g.Stage = stageSetup
	g.startTimer(int((setupTimePerName)/time.Second)*g.NamesPerPlayer, g.updatePlayers, g.startRound1)

	return nil
}

func (g *Game) switchTeams(who *Player) {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

	if g.Team1.removePlayer(who.Name) {
		g.Team2.addExistingPlayer(who)
	} else {
		g.Team2.removePlayer(who.Name)
		g.Team1.addExistingPlayer(who)
	}
}

func (g *Game) startRound1() {
	g.Lock()
	g.Stage = stageRound1
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()
}

// func (g *Game) loadNames() {

// }

func (g *Game) startTimer(seconds int, tick func(), finish func()) {
	go func() {
		g.Lock()
		g.Timer.Seconds = seconds
		g.Timer.Left = seconds
		g.Unlock()

		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			tick()
			g.Lock()
			g.Timer.Left--
			g.Unlock()
			if g.Timer.Left <= 0 {
				ticker.Stop()
			}
		}
		finish()
	}()
}
