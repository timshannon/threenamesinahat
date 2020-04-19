// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package server

// func errHandled(err error, w http.ResponseWriter, r *http.Request) bool {
// 	if err == nil {
// 		return false
// 	}

// 	var errMsg string
// 	var status int

// 	switch err.(type) {

// 	case *fail.Failure:
// 		errMsg = err.Error()
// 		status = err.(*fail.Failure).HTTPStatus
// 	case *json.SyntaxError, *json.UnmarshalTypeError:
// 		// Hardcoded external errors which can bubble up to the end users
// 		// without exposing internal server information, make them failures
// 		errMsg = fmt.Sprintf("We had trouble parsing your input, please check your input and try again: %s", err)
// 		status = http.StatusBadRequest
// 	default:
// 		status = http.StatusInternalServerError
// 		errMsg = fmt.Sprintf("An internal server error has occurred")
// 	}

// 	if strings.Contains(r.Header.Get("Accept"), "text/html") {
// 		w.WriteHeader(status)
// 		switch status {
// 		case http.StatusNotFound:
// 		default:
// 			log.Printf("HTTP Error: %s", err)
// 		}

// 	return true
// 	}
// 	// TODO: JSON? Websocket?  Need to determine how the UI is exactly interacting with the server first

// 	return true
// }
