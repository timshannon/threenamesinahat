// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package game

import (
	"time"
)

const timerPoll = 500 * time.Millisecond

func startTimer(duration time.Duration, tick func(passed time.Duration), finish, timeout func()) chan bool {
	stop := make(chan bool)

	go func() {
		c := time.After(duration)
		ticker := time.NewTicker(timerPoll)
		defer func() {
			ticker.Stop()
			if finish != nil {
				go finish()
			}
		}()

		last := time.Now()

		for {
			select {
			case <-stop:
				return
			case <-c:
				if timeout != nil {
					go timeout()
				}
				return
			case t := <-ticker.C:
				if tick != nil {
					go tick(t.Sub(last))
				}
				last = t
			}
		}
	}()
	return stop
}
