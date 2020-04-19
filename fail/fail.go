// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package fail

import (
	"fmt"
	"net/http"
)

// Failure is an error whos contents can be exposed to the client and is usually the result
// of incorrect client input
type Failure struct {
	Message    string `json:"message,omitempty"`
	HTTPStatus int    `json:"-"` //gets set in the error response
}

func (f *Failure) Error() string {
	return f.Message
}

// New creates a new failure with a default status of 400
func New(message string, args ...interface{}) *Failure {
	return &Failure{
		Message:    fmt.Sprintf(message, args...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewWithStatus creates a new failure with the passed in http status code
func NewWithStatus(message string, httpStatus int) *Failure {
	return &Failure{
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// FromErr returns a new failure based on the passed in error
// if passed in error is nil, then nil is returned
func FromErr(err error) *Failure {
	if err == nil {
		return nil
	}

	return NewWithStatus(err.Error(), http.StatusBadRequest)
}

// IsFailure tests whether the passed in error is a failure
func IsFailure(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *Failure:
		return true
	default:
		return false
	}
}

// NotFound creates a NotFound failure that returns to the user as a 404
func NotFound(message string, args ...interface{}) *Failure {
	return NewWithStatus(fmt.Sprintf(message, args...), http.StatusNotFound)
}

// Unauthorized returns an Unauthorized error for when a user doesn't have access to a resource
func Unauthorized(message string, args ...interface{}) *Failure {
	return NewWithStatus(fmt.Sprintf(message, args...), http.StatusUnauthorized)
}

// Conflict is the error retured when a record is being updated, but it's not the most current version
// of the record (409)
func Conflict(message string, args ...interface{}) *Failure {
	return NewWithStatus(fmt.Sprintf(message, args...), http.StatusConflict)
}
