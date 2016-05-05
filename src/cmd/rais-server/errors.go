package main

type HandlerError struct {
	Message string
	Code    int
}

func NewError(m string, c int) *HandlerError {
	return &HandlerError{m, c}
}
