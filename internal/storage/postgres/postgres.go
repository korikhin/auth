package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/studopolis/auth-server/internal/config"
	"github.com/studopolis/auth-server/internal/domain/models"
	"github.com/studopolis/auth-server/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(config config.Replica) (*Storage, error) {
	const op = "storage.postgres.New"

	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) User(ctx context.Context, id uint64) (*models.User, error) {
	const op = "storage.postgres.User"

	query := `
		select email, hash, role
		from public.users
		where id = $1;
	`
	user := &models.User{}
	err := s.pool.QueryRow(ctx, query, id).Scan(&user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.ID = id
	return user, nil
}

func (s *Storage) UserByCredentials(ctx context.Context, email string, hash []byte) (*models.User, error) {
	const op = "storage.postgres.UserByCredentials"

	query := `
		select id, role
		from public.users
		where email = $1 and hash = $2;
	`
	user := &models.User{}
	err := s.pool.QueryRow(ctx, query, email, hash).Scan(&user.ID, &user.Role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.Email = email
	user.PasswordHash = hash

	return user, nil
}

func (s *Storage) Ping(ctx context.Context) (string, error) {
	const op = "storage.postgres.Ping"

	var ping string

	query := `select version();`
	err := s.pool.QueryRow(ctx, query).Scan(&ping)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return ping, nil
}
