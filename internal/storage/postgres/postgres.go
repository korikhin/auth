package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/studopolis/auth-server/internal/config"
	"github.com/studopolis/auth-server/internal/domain/models"
	"github.com/studopolis/auth-server/internal/storage"

	codes "github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Options struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Storage struct {
	pool    *pgxpool.Pool
	mu      sync.Mutex
	Options Options
}

var (
	pool     *pgxpool.Pool
	poolOnce sync.Once
	poolErr  error
)

func New(ctx context.Context, config config.Storage) (*Storage, error) {
	const op = "storage.postgres.New"

	poolOnce.Do(func() {
		poolConfig, err := pgxpool.ParseConfig(config.Alpha.URL)
		if err != nil {
			poolErr = fmt.Errorf("%s: %w", op, err)
			return
		}

		poolConfig.MinConns = config.Alpha.MinConns
		poolConfig.MaxConns = config.Alpha.MaxConns
		poolConfig.MaxConnIdleTime = config.Alpha.IdleTimeout

		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			poolErr = fmt.Errorf("%s: %w", op, err)
			return
		}
	})

	if poolErr != nil {
		return nil, poolErr
	}

	opts := &Options{
		ReadTimeout:  config.Alpha.ReadTimeout,
		WriteTimeout: config.Alpha.WriteTimeout,
	}

	return &Storage{pool: pool, Options: *opts}, nil
}

func (s *Storage) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pool != nil {
		s.pool.Close()
		s.pool = nil
	}
}

func (s *Storage) User(ctx context.Context, id string) (*models.User, error) {
	const op = "storage.postgres.User"

	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	query := `
		select email, hash
		from public.users
		where id = @id;
	`
	args := pgx.NamedArgs{
		"id": userID,
	}

	user := &models.User{}
	err = s.pool.QueryRow(ctx, query, args).Scan(&user.Email, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.ID = id
	return user, nil
}

func (s *Storage) UserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "storage.postgres.UserByEmail"

	query := `
		select id, hash
		from public.users
		where email = @email;
	`
	args := pgx.NamedArgs{
		"email": email,
	}

	user := &models.User{}
	err := s.pool.QueryRow(ctx, query, args).Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.Email = email
	return user, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, hash []byte) (uint64, error) {
	const op = "storage.postgres.SaveUser"

	query := `
		insert into public.users(email, hash)
		values (@email, @hash)
		returning id;
	`
	args := pgx.NamedArgs{
		"email": email,
		"hash":  hash,
	}

	var userID uint64
	err := s.pool.QueryRow(ctx, query, args).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && codes.IsIntegrityConstraintViolation(pgErr.Code) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *Storage) Ping(ctx context.Context) error {
	const op = "storage.postgres.Ping"

	err := s.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
