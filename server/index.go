// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

import (
	"net/http"
)

func index(w *templateWriter, r *http.Request) {
	// if errHandled(fmt.Errorf("An error occurred"), w, r) {
	// 	return
	// }
	w.execute(struct {
		Name string
	}{
		Name: "Test Name",
	})
}
