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
	manager.Lock()
	defer manager.Unlock()

	rand.Seed(time.Now().UnixNano())
	g := &Game{
		state: GameState{
			Code:           generateCode(4),
			NamesPerPlayer: 3,
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
		if manager.games[i].Code() == code {
			return manager.games[i], true
		}
	}
	return nil, false
}

func Join(code, name string, fn MsgFunc) (*Player, error) {
	g, ok := Find(code)
	if !ok {
		return nil, fail.NotFound("Invalid Game code, try again")
	}
	player, err := g.join(name, fn)
	if err != nil {
		return nil, err
	}

	return player, nil
}

// TODO cleanup inactive games with no players
