package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/korikhin/auth/internal/config"
	"github.com/korikhin/auth/internal/domain/models"
	"github.com/korikhin/auth/internal/storage"

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
	poolOnce  sync.Once
	pool      *pgxpool.Pool
	initErr   error
	needsInit bool = true
	initMu    sync.Mutex
)

// TODO: Replace connection errors with custom error
//
// Prevent revealing of sensetive info, s.a. connection details etc.
func sanitizeError(err error) error {
	return err

	// var pgErr *pgconn.PgError
	// if !errors.As(err, &pgErr) {
	// 	return storage.ErrConnectionFailed
	// }
	// return err
}

func initializePool(ctx context.Context, config config.Storage) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, err
	}

	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MaxConnIdleTime = config.IdleTimeout

	newPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := newPool.Ping(ctx); err != nil {
		// TODO: Uncomment later
		// return nil, err
	}

	return newPool, nil
}

// New initializes a new Storage instance with a database connection pool
//
// It ensures a safe pool creation which persists across multiple
// calls until Stop is invoked.
func New(ctx context.Context, config config.Storage) (*Storage, error) {
	const op = "storage.postgres.New"

	initMu.Lock()
	defer initMu.Unlock()

	if needsInit {
		poolOnce.Do(func() {
			initCtx, cancel := context.WithTimeout(ctx, config.StartTimeout)
			defer cancel()

			pool, initErr = initializePool(initCtx, config)
			if initErr == nil {
				needsInit = false
			}
		})
	}

	if initErr != nil {
		return nil, fmt.Errorf("%s: %w", op, initErr)
	}

	opts := Options{
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &Storage{pool: pool, Options: opts}, nil
}

// Stop closes the connection pool and clears its resources.
// Call New to reinitialize the pool after calling Stop.
func (s *Storage) Stop() {
	initMu.Lock()
	defer initMu.Unlock()

	if !needsInit && s.pool != nil {
		s.pool.Close()
		s.pool = nil

		poolOnce = sync.Once{}
		needsInit = true
	}
}

// TODO: Test the methods below

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
		err = sanitizeError(err)
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
		err = sanitizeError(err)
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
		err = sanitizeError(err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *Storage) Ping(ctx context.Context) error {
	const op = "storage.postgres.Ping"

	err := s.pool.Ping(ctx)
	if err != nil {
		sanitizeError(err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
