// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/timshannon/threenamesinahat/fail"
)

type gameManager struct {
	sync.RWMutex
	games []*Game
}

var manager = &gameManager{}

const pollStatus = 60 * time.Second

var newGameRateDelay = &RateDelay{
	Type:   "newGame",
	Limit:  10,
	Delay:  5 * time.Second,
	Period: 5 * time.Minute,
	Max:    1 * time.Minute,
}

// New Starts a new game
func New(ipAddress string) (*Game, error) {
	_, err := newGameRateDelay.Attempt(ipAddress)
	if err != nil {
		return nil, err
	}
	rand.Seed(time.Now().UnixNano())
	code := generateCode(4)
	manager.Lock()
	defer manager.Unlock()

	g := &Game{
		gameState: gameState{
			Code:           code,
			NamesPerPlayer: 3,
			Stage:          stagePregame,
		},
	}
	reset(g, "")

	time.AfterFunc(pollStatus, func() { cleanGame(g) })

	manager.games = append(manager.games, g)
	return g, nil
}

func cleanGame(g *Game) {
	if g.IsDead() {
		removeGame(g)
		return
	}
	time.AfterFunc(pollStatus, func() { cleanGame(g) })
}

// Find finds a game by the given Code
func Find(code string) (*Game, bool) {
	manager.RLock()
	defer manager.RUnlock()
	code = strings.ToUpper(code)

	for i := range manager.games {
		if manager.games[i].Code == code {
			return manager.games[i], true
		}
	}
	return nil, false
}

// Join allows a player to join a game in progress
func Join(code, name string) (*Player, error) {
	g, ok := Find(code)
	if !ok {
		return nil, fail.NotFound("Invalid Game code, try again")
	}
	player, err := g.join(name)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func removeGame(g *Game) {
	manager.Lock()
	defer manager.Unlock()

	for i := range manager.games {
		if manager.games[i].Code == g.Code {
			manager.games = append(manager.games[:i], manager.games[i+1:]...)
			log.Printf("Removing game %s", g.Code)
			return
		}
	}
}
