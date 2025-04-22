package service

import (
	"context"
	"time"

	"github.com/gurebusan/simple-auth/internal/config"
	"github.com/gurebusan/simple-auth/internal/models"
	"github.com/gurebusan/simple-auth/internal/storage"
)

type Storage interface {
	SaveRefreshToken(ctx context.Context, token models.Token) error
	FindRefreshToken(ctx context.Context, guid string) (token models.Token, err error)
	RemoveRefreshToken(ctx context.Context, guid string) error
	ReplaceRefreshToken(ctx context.Context, token models.Token) error
}

type TokenManager interface {
	NewAccessToken(guid string, ttl time.Duration) (string, error)
	NewRefreshToken() (string, error)
	HashToken(token string) ([]byte, error)
	CompareHash(providedToken string, hashed []byte) bool
}

type Notifier interface {
	Send(guid, email, oldIP, newIP string) error
}

type Service struct {
	ctx          context.Context
	storage      Storage
	tokenManager TokenManager
	notifier     Notifier
	cfg          *config.Config
}

func New(ctx context.Context, storage Storage, tokenManager TokenManager, notifer Notifier, cfg *config.Config) *Service {
	return &Service{
		ctx:          ctx,
		storage:      storage,
		tokenManager: tokenManager,
		notifier:     notifer,
		cfg:          cfg,
	}
}

func (s *Service) IssueTokens(guid, email, ip string) (accessToken, refreshToken string, err error) {
	accessToken, err = s.tokenManager.NewAccessToken(guid, s.cfg.Token.AccessTTL)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return "", "", err
	}

	hashedToken, err := s.tokenManager.HashToken(refreshToken)
	if err != nil {
		return "", "", err
	}
	token := models.NewToken(guid, email, ip, hashedToken)
	err = s.storage.SaveRefreshToken(s.ctx, token)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *Service) RefreshTokens(guid, currentIp, oldRefreshToken string) (newAccessToken, newRefreshToken string, err error) {
	token, err := s.storage.FindRefreshToken(s.ctx, guid)
	if err != nil {
		return "", "", err
	}

	if token.IP != currentIp {
		_ = s.notifier.Send(guid, token.Email, token.IP, currentIp) // Моковая реализация, игнорируем ошибку, логируем
	}

	if !s.tokenManager.CompareHash(oldRefreshToken, token.Hash) {
		return "", "", storage.ErrTokenNotFound
	}

	if time.Since(token.CreatedAt) > s.cfg.Token.RefreshTTL {
		s.storage.RemoveRefreshToken(s.ctx, guid)
		return "", "", storage.ErrTokenExpired
	}

	newRefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return "", "", err
	}

	newHashedToken, err := s.tokenManager.HashToken(newRefreshToken)
	if err != nil {
		return "", "", err
	}
	newToken := models.NewToken(guid, token.Email, token.IP, newHashedToken)
	err = s.storage.ReplaceRefreshToken(s.ctx, newToken)
	if err != nil {
		return "", "", err
	}

	newAccessToken, err = s.tokenManager.NewAccessToken(guid, s.cfg.Token.AccessTTL)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
