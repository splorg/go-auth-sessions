package handler

import "net/http"

type APIError struct {
	StatusCode int
	Err        error
}

func (e *APIError) Error() string {
	return e.Err.Error()
}

func NewAPIError(statusCode int, err error) *APIError {
	return &APIError{StatusCode: statusCode, Err: err}
}

type ApiHandler func(http.ResponseWriter, *http.Request) *APIError
