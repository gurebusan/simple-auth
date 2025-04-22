package storage

import "errors"

var (
	ErrTokenNotFound = errors.New("refresh token not found or already used")
	ErrTokenExpired  = errors.New("refresh token expired and has been deleted")
	ErrUserNotFound  = errors.New("user not found")
)
