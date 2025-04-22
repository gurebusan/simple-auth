package models

import "time"

type Token struct {
	GUID      string
	Email     string
	IP        string
	Hash      []byte
	CreatedAt time.Time
}

func NewToken(guid, email, ip string, hash []byte) Token {
	return Token{
		GUID:      guid,
		Email:     email,
		IP:        ip,
		Hash:      hash,
		CreatedAt: time.Now().UTC(),
	}
}
