package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/gurebusan/simple-auth/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseDSN string) (*Storage, error) {
	const op = "storage.New.pgxpool.New"
	pool, err := pgxpool.New(ctx, databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Storage) SaveRefreshToken(ctx context.Context, token models.Token) error {
	const op = "storage.SaveRefreshToken"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)
	query := `INSERT INTO tokens (guid, email, ip, hash, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(ctx, query, token.GUID, token.Email, token.IP, token.Hash, token.CreatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return s.ReplaceRefreshToken(ctx, token)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) FindRefreshToken(ctx context.Context, guid string) (token models.Token, err error) {
	const op = "storage.FindRefreshToken"
	query := `SELECT guid, email, ip, hash, created_at FROM tokens WHERE guid = $1`
	err = s.pool.QueryRow(ctx, query, guid).Scan(&token.GUID, &token.Email, &token.IP, &token.Hash, &token.CreatedAt)
	if err != nil {
		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

func (s *Storage) RemoveRefreshToken(ctx context.Context, guid string) error {
	const op = "storage.RemoveRefreshToken"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)
	query := `DELETE FROM tokens WHERE guid = $1`
	_, err = tx.Exec(ctx, query, guid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) ReplaceRefreshToken(ctx context.Context, token models.Token) error {
	const op = "storage.ReplaceRefreshToken"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)
	query := `UPDATE tokens SET hash = $1, ip = $2, created_at = $3 WHERE guid = $4`
	_, err = tx.Exec(ctx, query, token.Hash, token.IP, token.CreatedAt, token.GUID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
