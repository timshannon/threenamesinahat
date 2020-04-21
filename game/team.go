// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

type Team struct {
	name    string
	color   string
	players []Player
}

func (t *Team) player(name string) (*Player, bool) {
	for i := range t.players {
		if t.players[i].name == name {
			return &t.players[i], true
		}
	}
	return nil, false
}
