package manager

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	secret string
}

func NewTokenManager(secret string) *Manager {
	return &Manager{secret: secret}
}

func (m *Manager) NewAccessToken(guid string, ttl time.Duration) (string, error) {
	const op = "manager.NewAccessToken"
	token := jwt.New(jwt.SigningMethodHS512)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = guid
	claims["exp"] = time.Now().Add(ttl).Unix()
	claims["iat"] = time.Now().Unix()
	res, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return res, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	const op = "manager.NewRefreshToken"
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (m *Manager) HashToken(token string) ([]byte, error) {
	const op = "manager.HashToken"
	res, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return res, nil
}

func (m *Manager) CompareHash(providedToken string, hashed []byte) bool {
	return bcrypt.CompareHashAndPassword(hashed, []byte(providedToken)) == nil
}
