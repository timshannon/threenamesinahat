// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"math/rand"
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
	stagePregame  = "pregame" // players join
	stageSetup    = "setup"   // players add names
	stagePlaying  = "playing"
	stageStealing = "stealing"
	stageEnd      = "end"
)

const (
	secondsPerTurn      = 30 // how much time each player gets per turn
	setupSecondsPerName = 30 // how much time per name each player gets during game setup
	secondsToSteal      = 15 // how much time the opposing team gets to steal
)

type Game struct {
	sync.RWMutex // manage the lock in methods, functions expect lock to be already managed
	gameState
}

// gameState is copied out for JSON encoding
// mutexes can't be copied, so sync access is managed in Game
// RWMutex is Read Locked, then the state is copied to send to players
type gameState struct {
	Code           string  `json:"code"`
	NamesPerPlayer int     `json:"namesPerPlayer"`
	Team1          Team    `json:"team1"`
	Team2          Team    `json:"team2"`
	Leader         *Player `json:"leader"`
	Stage          string  `json:"stage"`
	Round          int     `json:"round"`
	Timer          struct {
		Seconds      int `json:"seconds"`
		Left         int `json:"left"`
		durationLeft time.Duration
	} `json:"timer"`
	ClueGiver *Player `json:"clueGiver"`

	nameList       []string
	clueGiverTrack struct {
		team1Index int
		team2Index int
		team1      bool
	}
	canSteal bool
	Stats    struct {
		Winner     int `json:"winner"`
		Team1Score int `json:"team1Score"`
		Team2Score int `json:"team2Score"`
	} `json:"stats"`
}

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
		if g.Leader == nil && len(g.Team1.Players) == 1 {
			// first player in is leader
			g.Leader = player
		}

		return player, nil
	}

	player := g.Team2.addNewPlayer(name, g)
	return player, nil
}

func (g *Game) setNamesPerPlayer(who *Player, num int) error {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

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

	g.NamesPerPlayer = num
	return nil
}

func (g *Game) updatePlayers() {
	g.RLock()
	defer g.RUnlock()
	g.Team1.updatePlayers(g.gameState)
	g.Team2.updatePlayers(g.gameState)
}

// same as method, except game lock is already managed
func updatePlayers(g *Game) {
	g.Team1.updatePlayers(g.gameState)
	g.Team2.updatePlayers(g.gameState)
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

	cleanPlayers(g)

	if len(g.Team1.Players) < 2 || len(g.Team2.Players) < 2 {
		// return to pregame and wait for players to join
		return nil
	}

	g.Stage = stageSetup
	g.startTimer(setupSecondsPerName*g.NamesPerPlayer, func() {
		g.RLock()
		startRound := true
		for _, p := range g.Team1.Players {
			if len(p.names()) < g.NamesPerPlayer {
				startRound = false
				break
			}
		}

		if startRound {
			for _, p := range g.Team2.Players {
				if len(p.names()) < g.NamesPerPlayer {
					startRound = false
					break
				}
			}
		}
		g.RUnlock()
		if startRound {
			// if all players have submitted the necessary names, end the timer early and start the round
			g.stopTimer() // will start the round on the subsequent tick
		}
		g.updatePlayers()
	}, func() { g.startRound(1) })

	return nil
}

func (g *Game) cleanPlayers() {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()
	cleanPlayers(g)
}

func (g *Game) IsDead() bool {
	g.RLock()
	defer g.RUnlock()
	return len(g.Team1.Players) == 0 && len(g.Team2.Players) == 0
}

func cleanPlayers(g *Game) {
	g.Team1.cleanPlayers()
	g.Team2.cleanPlayers()
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

func (g *Game) stopTimer() {
	g.Lock()
	g.Timer.Left = 0
	g.Timer.durationLeft = 0
	updatePlayers(g)
	g.Unlock()
}

func (g *Game) startTimer(seconds int, tick func(), finish func()) {
	go func() {
		g.Lock()
		g.Timer.Seconds = seconds
		g.Timer.Left = seconds
		g.Timer.durationLeft = time.Duration(g.Timer.Left * int(time.Second))
		updatePlayers(g)
		g.Unlock()

		poll := 500 * time.Millisecond

		ticker := time.NewTicker(poll)
		for range ticker.C {
			if tick != nil {
				tick()
			}
			g.Lock()
			g.Timer.durationLeft = g.Timer.durationLeft - poll
			g.Timer.Left = int(g.Timer.durationLeft / time.Second)
			if g.Timer.Left <= 0 {
				ticker.Stop()
				g.Unlock()
				break
			}
			g.Unlock()
		}

		if finish != nil {
			finish()
		}
	}()
}

func (g *Game) startRound(round int) {
	if round != 1 {
		g.stopTimer() // stop timer incase previous round end early
	}
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

	g.Stage = stagePlaying
	g.Round = round
	g.canSteal = false
	loadNames(g)
	if round == 1 {
		// reset player list
		g.clueGiverTrack.team1 = false
		g.clueGiverTrack.team1Index = -1
		g.clueGiverTrack.team2Index = -1
		nextPlayerTurn(g)
	}
}

func shuffleNames(g *Game) {
	rand.Shuffle(len(g.nameList), func(i, j int) {
		g.nameList[i], g.nameList[j] = g.nameList[j], g.nameList[i]
	})
}

func loadNames(g *Game) {
	g.nameList = nil

	for _, p := range g.Team1.Players {
		g.nameList = append(g.nameList, p.names()...)
	}

	for _, p := range g.Team2.Players {
		g.nameList = append(g.nameList, p.names()...)
	}

	shuffleNames(g)
}

func nextPlayerTurn(g *Game) {
	g.Stage = stagePlaying
	shuffleNames(g)
	g.clueGiverTrack.team1 = !g.clueGiverTrack.team1
	if g.clueGiverTrack.team1 {
		g.clueGiverTrack.team1Index++
		if g.clueGiverTrack.team1Index >= len(g.Team1.Players) {
			g.clueGiverTrack.team1Index = 0
		}
		g.ClueGiver = g.Team1.Players[g.clueGiverTrack.team1Index]
		return
	}

	g.clueGiverTrack.team2Index++
	if g.clueGiverTrack.team2Index >= len(g.Team2.Players) {
		g.clueGiverTrack.team2Index = 0
	}
	g.ClueGiver = g.Team2.Players[g.clueGiverTrack.team2Index]
}

func (g *Game) startTurn(p *Player) error {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

	if g.Stage != stagePlaying {
		return nil
	}

	if g.ClueGiver == nil || g.ClueGiver.Name != p.Name {
		return nil
	}

	g.canSteal = true
	g.startTimer(secondsPerTurn, g.updatePlayers, func() {
		g.Lock()
		defer func() {
			g.Unlock()
			g.updatePlayers()
		}()
		if g.canSteal {
			steal(g)
		} else {
			nextPlayerTurn(g)
		}
	})

	if len(g.nameList) == 0 {
		g.ClueGiver = nil
		if g.Round == 3 {
			go g.endGame() // run on a separate go routine to prevent deadlock
			return nil
		}
		go g.startRound(g.Round + 1) // run on a separate go routine to prevent deadlock

		return nil
	}
	p.Send <- Msg{Type: "name", Data: p.game.nameList[0]}

	return nil
}

func (g *Game) nextName(p *Player) error {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

	if g.Stage != stagePlaying {
		return nil
	}

	if g.ClueGiver == nil || g.ClueGiver.Name != p.Name {
		return nil
	}

	if g.Timer.Left == 0 {
		return nil
	}

	g.nameList = g.nameList[1:]
	if g.clueGiverTrack.team1 {
		g.Stats.Team1Score++
	} else {
		g.Stats.Team2Score++
	}
	if len(g.nameList) == 0 {
		g.ClueGiver = nil
		if g.Round == 3 {
			go g.endGame() // run on a separate go routine to prevent deadlock
			return nil
		}
		go g.startRound(g.Round + 1) // run on a separate go routine to prevent deadlock

		return nil
	}

	p.Send <- Msg{Type: "name", Data: g.nameList[0]}
	return nil
}

// send final answer vote button to stealing team
// if entire team responds final answer before timer runs out, then ClueGiver gets to
// set if they got it right or not
func steal(g *Game) {
	if g.Stage != stagePlaying {
		return
	}
	g.Stage = stageStealing

	g.startTimer(secondsToSteal, g.updatePlayers, nil)

	g.ClueGiver.Send <- Msg{Type: "stealcheck"}
}

func (g *Game) stealConfirm(p *Player, correct bool) error {
	g.stopTimer()
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()
	if g.Stage != stageStealing {
		return fail.New("Turn is not being stolen currently")
	}

	if g.ClueGiver == nil || g.ClueGiver.Name != p.Name {
		return nil
	}

	if correct {
		g.nameList = g.nameList[1:]
		if g.clueGiverTrack.team1 {
			g.Stats.Team2Score++
		} else {
			g.Stats.Team1Score++
		}

		if len(g.nameList) == 0 {
			if g.Round == 3 {
				go g.endGame() // run on a separate go routine to prevent deadlock
				return nil
			}
			go g.startRound(g.Round + 1) // run on a separate go routine to prevent deadlock

			return nil
		}
		nextPlayerTurn(g)
		return nil
	}

	nextPlayerTurn(g)
	return nil
}

func (g *Game) endGame() {
	g.stopTimer()
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()
	g.Stage = stageEnd
	if g.Stats.Team1Score > g.Stats.Team2Score {
		g.Stats.Winner = 1
	} else if g.Stats.Team2Score > g.Stats.Team1Score {
		g.Stats.Winner = 2
	}

}
