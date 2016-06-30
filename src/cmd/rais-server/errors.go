package main

// HandlerError represents an HTTP error message and status code
type HandlerError struct {
	Message string
	Code    int
}

// NewError generates a new HandlerError with the given message and code
func NewError(m string, c int) *HandlerError {
	return &HandlerError{m, c}
}
