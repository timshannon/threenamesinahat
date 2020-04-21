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

type Game struct {
	sync.RWMutex
	code     string
	numNames int
	team1    Team
	team2    Team
}

// Code returns the games code
func (g *Game) Code() string { return g.code }

// Join adds a new player to the game
func (g *Game) Join(name string) (*Player, error) {
	g.Lock()
	defer g.Unlock()
	if player, ok := g.team1.player(name); ok {
		if player.connected {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		return player, nil
	}

	if player, ok := g.team2.player(name); ok {
		if player.connected {
			return nil, fail.New("A player with the name " + name + " is already connected, please choose a new name")
		}
		return player, nil
	}

	// new player
	player := Player{name: name}
	if len(g.team1.players) <= len(g.team2.players) {
		g.team1.players = append(g.team1.players, player)
		return &player, nil
	}

	g.team2.players = append(g.team2.players, player)
	return &player, nil
}

// generateCode generates a random string of only uppercase letters (A-Z) of the specified length
func generateCode(length int) string {
	code := ""

	for i := 0; i < length; i++ {
		code += string(codeChars[rand.Intn(26)])
	}

	return code
}
