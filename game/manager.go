// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
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

// New Starts a new game
func New() *Game {
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

	manager.games = append(manager.games, g)
	return g
}

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

// TODO cleanup inactive games with no players
