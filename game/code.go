// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"math/rand"
	"strings"
)

const codeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generateCode generates a random string of only uppercase letters (A-Z) of the specified length
func generateCode(length int) string {
	code := ""

	for i := 0; i < length; i++ {
		code += string(codeChars[rand.Intn(len(codeChars))])
	}

	if isDirty(code) {
		return generateCode(length)
	}

	// make sure code isn't already in use
	if _, ok := Find(code); ok {
		return generateCode(length)
	}

	return code
}

func isDirty(str string) bool {
	for i := range dirtyWords {
		if strings.Contains(str, dirtyWords[i]) {
			return true
		}
	}
	return false
}

// Comment out this list for the worlds slowest drinking game...
var dirtyWords = []string{
	"SHIT",
	"FUCK",
	"DAMN",
	"CUNT",
	"TIT",
	"PISS",
	"BOOB",
	"BOOB", // reversed just in case
	"ASS",
	"CUM",
	"FAG",
	"ANAL",
	"ANUS",
	"ARSE",
	"CLIT",
	"COCK",
	"CRAP",
	"DICK",
	"DUMB",
	"DYKE",
	"GOOK",
	"HOMO",
	"JISM",
	"JUGS",
	"KIKE",
	"PAKI",
	"PISS",
	"SCUM",
	"SHAG",
	"SLUT",
	"SPIC",
	"SUCK",
	"TURD",
	"TWAT",
	"WANK",
	// feel free to open a PR to add more. I promise I won't judge
}
