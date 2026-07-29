package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyInUse  = errors.New("email already in use")
	ErrInvalidInput		  = errors.New("invalid input")	 
)