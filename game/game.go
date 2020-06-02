// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"encoding/json"
	"fmt"
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
	stagePregame     = "pregame" // players join
	stageSetup       = "setup"   // players add names
	stagePlaying     = "playing"
	stageRoundChange = "roundchange"
	stageStealing    = "stealing"
	stageEnd         = "end"
)

const (
	secondsPerTurn      = 30 // how much time each player gets per turn
	setupSecondsPerName = 30 // how much time per name each player gets during game setup
	secondsToSteal      = 15 // how much time the opposing team gets to steal
)

const (
	soundTick       = "tick"
	soundTimerAlarm = "timer-alarm"
	soundScore      = "score"
	soundNotify     = "notify"
	soundRoundEnd   = "round-end"
	soundGameWin    = "game-win"
	soundGameLose   = "game-lose"
)

type Game struct {
	sync.RWMutex // manage the lock in methods, functions expect lock to be already managed
	gameState
	rand *rand.Rand
}

type nameItem struct {
	name   string
	player string
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
		stop         chan bool
	} `json:"timer"`
	ClueGiver *Player `json:"clueGiver"`

	nameList []nameItem

	clueGiverTrack struct {
		team1Index int
		team2Index int
		team1      bool
	}
	canSteal bool
	Stats    struct {
		Winner        int `json:"winner"`
		Team1Score    int `json:"team1Score"`
		Team2Score    int `json:"team2Score"`
		BestClueGiver struct {
			Player  string `json:"player"`
			Guesses int    `json:"guesses"`
			stats   map[string]int
		} `json:"bestClueGiver"` // who earned the most guess when giving clues
		MostStolen struct {
			Player string `json:"player"`
			Steals int    `json:"steals"`
			stats  map[string]int
		} `json:"mostStolen"` // who had the most names stolen when it was their turn
		EasiestName struct {
			Name      string `json:"name"`
			Submitter string `json:"submitter"`
			GuessTime string `json:"guessTime"`
			guessTime time.Duration
		} `json:"easiestName"` // which name was guessed the fastest
		HardestName struct {
			Name      string `json:"name"`
			Submitter string `json:"submitter"`
			GuessTime string `json:"guessTime"`
			guessTime time.Duration
		} `json:"hardestName"` // which name took the longest to guess
		nameTime time.Time
	} `json:"stats"`
}

func (g *Game) MarshalJSON() ([]byte, error) {
	g.RLock()
	defer g.RUnlock()
	return json.Marshal(g.gameState)
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
	state := g.gameState

	g.Team1.updatePlayers(state)
	g.Team2.updatePlayers(state)
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
		playTimerSound(g, &g.Team1)
		playTimerSound(g, &g.Team2)

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
	}, func() {
		g.RLock()
		// don't start the round if no one submitted names in time
		startRound := false
		for _, p := range g.Team1.Players {
			if len(p.names()) > 0 {
				startRound = true
				break
			}
		}
		if !startRound {
			for _, p := range g.Team2.Players {
				if len(p.names()) > 0 {
					startRound = true
					break
				}
			}
		}
		g.RUnlock()
		if !startRound {
			g.Lock()
			g.Stage = stagePregame
			g.Unlock()
			g.updatePlayers()
			return
		}
		g.changeRound(1)
	}, func() {
		g.Team1.playSound(soundTimerAlarm)
		g.Team2.playSound(soundTimerAlarm)
	})

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

	if len(g.Team1.Players) < 2 || len(g.Team2.Players) < 2 {
		if g.Stage != stagePregame && g.Stage != stageEnd {
			reset(g, "Not enough players to continue")
			updatePlayers(g)
		}
		return
	}

	if !g.Leader.ping() {
		g.Leader = g.Team1.Players[0]
	}
}

func (g *Game) removePlayer(name string) {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()
	if !g.Team1.removePlayer(name) {
		g.Team2.removePlayer(name)
	}

	if len(g.Team1.Players) < 2 || len(g.Team2.Players) < 2 {
		if g.Stage != stagePregame && g.Stage != stageEnd {
			reset(g, "Not enough players to continue")
			updatePlayers(g)
		}
		return
	}
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
	stopTimer(g)
	g.Unlock()
}

func stopTimer(g *Game) {
	if g.Timer.stop != nil {
		g.Timer.stop <- true
		g.Timer.Left = 0
		g.Timer.durationLeft = 0
		g.Timer.stop = nil
	}
}

func playTimerSound(g *Game, team *Team) {
	ratio := float64(g.Timer.Left) / float64(g.Timer.Seconds)
	if ratio <= .25 {
		// tick every half second
		team.playSound(soundTick)
		return
	}
	if ratio <= .5 {
		if g.Timer.Left == int(g.Timer.durationLeft.Round(time.Second)/time.Second) {
			// tick every second
			team.playSound(soundTick)
		}
	}
}

func (g *Game) startTimer(seconds int, tick func(), finish func(), timeout func()) {
	go func() {
		g.Lock()
		defer func() {
			g.Unlock()
			g.updatePlayers()
		}()

		g.Timer.Seconds = seconds
		g.Timer.Left = seconds
		g.Timer.durationLeft = time.Duration(g.Timer.Left * int(time.Second))

		g.Timer.stop = startTimer(g.Timer.durationLeft, func(passed time.Duration) {
			g.Lock()
			g.Timer.durationLeft -= passed
			g.Timer.Left = int(g.Timer.durationLeft / time.Second)
			g.Unlock()
			if tick != nil {
				tick()
			}
		}, func() {
			g.Lock()
			g.Timer.stop = nil
			g.Unlock()
			if finish != nil {
				finish()
			}
		}, timeout)
	}()
}

func (g *Game) changeRound(round int) {
	g.stopTimer() // stop timer incase previous round end early
	g.Lock()
	defer g.Unlock()

	cleanPlayers(g)
	if g.Stage == stagePregame {
		// game got reset
		return
	}

	g.Stage = stageRoundChange
	updatePlayers(g)
	g.Team1.playSound(soundRoundEnd)
	g.Team2.playSound(soundRoundEnd)

	g.startTimer(10, g.updatePlayers, func() {
		g.startRound(round)
	}, nil)
}

func (g *Game) startRound(round int) {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

	g.Stage = stagePlaying
	g.Round = round
	g.canSteal = false
	loadNames(g)
	nextPlayerTurn(g)
}

func shuffleNames(g *Game) {
	g.rand.Seed(time.Now().UnixNano())
	g.rand.Shuffle(len(g.nameList), func(i, j int) {
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
	defer func() {
		g.ClueGiver.playSound(soundNotify)
		g.ClueGiver.SendMsg(Msg{Type: "startcheck"})
	}()

	cleanPlayers(g)
	if g.Stage == stagePregame {
		// game got reset
		return
	}

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
	team := &g.Team1
	if !g.clueGiverTrack.team1 {
		team = &g.Team2
	}
	g.startTimer(secondsPerTurn, func() {
		g.RLock()
		playTimerSound(g, team)
		g.RUnlock()
		g.updatePlayers()
	}, func() {
		if g.canSteal {
			g.steal()
		} else {
			g.Lock()
			nextPlayerTurn(g)
			g.updatePlayers()
			g.Unlock()
		}
	}, func() {
		team.playSound(soundTimerAlarm)
	})

	if len(g.nameList) == 0 {
		return nil
	}
	g.Stats.nameTime = time.Now()
	p.SendMsg(Msg{Type: "name", Data: p.game.nameList[0].name})

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

	updateNameStats(g, false)

	g.nameList = g.nameList[1:]
	if g.clueGiverTrack.team1 {
		g.Stats.Team1Score++
		g.Team1.playSound(soundScore)
	} else {
		g.Stats.Team2Score++
		g.Team2.playSound(soundScore)
	}

	if len(g.nameList) == 0 {
		g.ClueGiver = nil
		if g.Round == 3 {
			go g.endGame() // run on a separate go routine to prevent deadlock
			return nil
		}
		go g.changeRound(g.Round + 1) // run on a separate go routine to prevent deadlock

		return nil
	}

	p.SendMsg(Msg{Type: "name", Data: g.nameList[0].name})
	return nil
}

func updateNameStats(g *Game, steal bool) {
	if steal {
		g.Stats.MostStolen.stats[g.ClueGiver.Name]++
	} else {
		g.Stats.BestClueGiver.stats[g.ClueGiver.Name]++
	}
	diff := time.Now().Sub(g.Stats.nameTime)
	name := g.nameList[0]

	if diff > g.Stats.HardestName.guessTime {
		g.Stats.HardestName.guessTime = diff
		g.Stats.HardestName.Name = name.name
		g.Stats.HardestName.Submitter = name.player
		g.Stats.HardestName.GuessTime = fmt.Sprintf("%9.1f seconds", diff.Round(time.Millisecond).Seconds())
	}
	if diff < g.Stats.EasiestName.guessTime || g.Stats.EasiestName.guessTime == 0 {
		g.Stats.EasiestName.guessTime = diff
		g.Stats.EasiestName.Name = name.name
		g.Stats.EasiestName.Submitter = name.player
		g.Stats.EasiestName.GuessTime = fmt.Sprintf("%9.1f seconds", diff.Round(time.Millisecond).Seconds())
	}
}

// send final answer vote button to stealing team
// if entire team responds final answer before timer runs out, then ClueGiver gets to
// set if they got it right or not
func (g *Game) steal() {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

	if g.Stage != stagePlaying {
		return
	}
	g.Stage = stageStealing

	g.Unlock()
	// update to steal stage immediately
	g.updatePlayers()
	g.Lock()
	team := &g.Team1
	if g.clueGiverTrack.team1 {
		team = &g.Team2
	}

	g.startTimer(secondsToSteal, func() {
		g.RLock()
		playTimerSound(g, team)
		g.RUnlock()
		g.updatePlayers()
	}, nil, func() {
		g.Lock()
		defer func() {
			g.Unlock()
			g.updatePlayers()
		}()
		team.playSound(soundTimerAlarm)
		nextPlayerTurn(g)
	})

	g.ClueGiver.SendMsg(Msg{Type: "stealcheck"})
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
		updateNameStats(g, true)
		g.nameList = g.nameList[1:]
		if g.clueGiverTrack.team1 {
			g.Stats.Team2Score++
			g.Team2.playSound(soundScore)
		} else {
			g.Stats.Team1Score++
			g.Team1.playSound(soundScore)
		}

		if len(g.nameList) == 0 {
			if g.Round == 3 {
				go g.endGame() // run on a separate go routine to prevent deadlock
				return nil
			}
			go g.changeRound(g.Round + 1) // run on a separate go routine to prevent deadlock

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
		g.Team1.playSound(soundGameWin)
		g.Team2.playSound(soundGameLose)
	} else if g.Stats.Team2Score > g.Stats.Team1Score {
		g.Stats.Winner = 2
		g.Team2.playSound(soundGameWin)
		g.Team1.playSound(soundGameLose)
	}

	for player, guesses := range g.Stats.BestClueGiver.stats {
		if guesses > g.Stats.BestClueGiver.Guesses {
			g.Stats.BestClueGiver.Guesses = guesses
			g.Stats.BestClueGiver.Player = player
		}
	}

	for player, steals := range g.Stats.MostStolen.stats {
		if steals > g.Stats.MostStolen.Steals {
			g.Stats.MostStolen.Steals = steals
			g.Stats.MostStolen.Player = player
		}
	}

}

func (g *Game) reset(p *Player, reason string) error {
	g.Lock()
	defer func() {
		g.Unlock()
		g.updatePlayers()
	}()

	if g.Stage != stageEnd {
		return fail.New("Game has not yet ended")
	}

	if !p.isLeader() {
		return fail.New("Only the game leader can reset the game")
	}
	reset(g, reason)
	return nil
}

func reset(g *Game, reason string) {
	stopTimer(g)
	g.Stage = stagePregame
	g.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	g.Round = 0
	g.ClueGiver = nil
	g.clueGiverTrack.team1 = false
	g.clueGiverTrack.team1Index = -1
	g.clueGiverTrack.team2Index = -1
	g.nameList = nil
	g.Team1.clearNames()
	g.Team2.clearNames()
	g.canSteal = false
	g.Stats.Winner = 0
	g.Stats.Team1Score = 0
	g.Stats.Team2Score = 0
	g.Stats.BestClueGiver.Player = ""
	g.Stats.BestClueGiver.Guesses = 0
	g.Stats.BestClueGiver.stats = make(map[string]int)
	g.Stats.MostStolen.Player = ""
	g.Stats.MostStolen.Steals = 0
	g.Stats.MostStolen.stats = make(map[string]int)
	g.Stats.EasiestName.Name = ""
	g.Stats.EasiestName.Submitter = ""
	g.Stats.EasiestName.GuessTime = ""
	g.Stats.HardestName.Name = ""
	g.Stats.HardestName.Submitter = ""
	g.Stats.HardestName.GuessTime = ""

	if reason != "" {
		g.Team1.sendNotification(reason)
		g.Team2.sendNotification(reason)
	}
}
