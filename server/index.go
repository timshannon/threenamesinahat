package server

import "net/http"

func index(w *templateWriter, r *http.Request) {
	w.execute(struct {
		Name string
	}{
		Name: "Test Name",
	})
}
