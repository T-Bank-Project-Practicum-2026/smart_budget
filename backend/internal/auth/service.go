package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"smart_budget/internal/db/apidb"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email" format:"email"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"createdAt" format:"date-time"`
}

type Session struct {
	Access    string
	Refresh   string
	ExpiresAt time.Time
}

type ActiveSession struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}

var ErrAccountChanged = errors.New("account_changed")

func (s *Service) ActiveSession(ctx context.Context, id, userID string) (ActiveSession, error) {
	row, err := apidb.New(s.db).GetActiveSession(ctx, apidb.GetActiveSessionParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return ActiveSession{}, ErrUnauthorized
	}

	return ActiveSession{ID: row.ID, UserID: row.UserID, ExpiresAt: row.ExpiresAt}, err
}

func (s *Service) Identity(ctx context.Context, access string) (ActiveSession, error) {
	claims, err := s.tokens.Parse(access, time.Now())
	if err != nil {
		return ActiveSession{}, err
	}

	return s.ActiveSession(ctx, claims.SessionID, claims.Subject)
}

type LoginResult struct {
	User    User
	Session Session
}

type Service struct {
	db        *pgxpool.Pool
	tokens    *Tokens
	passwords passwords
}

func NewService(db *pgxpool.Pool, tokens *Tokens) (*Service, error) {
	passwords, err := newPasswords()
	if err != nil {
		return nil, err
	}

	return &Service{db: db, tokens: tokens, passwords: passwords}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (LoginResult, error) {
	if err := input.validate(); err != nil {
		return LoginResult{}, err
	}

	if err := s.allowAttempt(ctx, input.Email); err != nil {
		return LoginResult{}, err
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		return LoginResult{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LoginResult{}, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	q := apidb.New(tx)
	row, err := q.CreateUser(ctx, apidb.CreateUserParams{ID: rand.Text(), Email: input.Email, PasswordHash: hash, Timezone: input.Timezone})
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return LoginResult{}, ErrAccountExists
		}

		return LoginResult{}, err
	}

	session, err := s.newSession(ctx, q, row.ID)
	if err != nil {
		return LoginResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{User: User{ID: row.ID, Email: row.Email, Timezone: row.Timezone, CreatedAt: row.CreatedAt}, Session: session}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	if err := input.validate(); err != nil {
		return LoginResult{}, err
	}

	if err := s.allowAttempt(ctx, input.Email); err != nil {
		return LoginResult{}, err
	}

	q := apidb.New(s.db)
	row, err := q.FindUserByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return LoginResult{}, err
	}

	if !s.passwords.check(row.PasswordHash, input.Password) {
		return LoginResult{}, ErrInvalidCredentials
	}

	session, err := s.newSession(ctx, q, row.ID)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{User: User{ID: row.ID, Email: row.Email, Timezone: row.Timezone, CreatedAt: row.CreatedAt}, Session: session}, nil
}

func (s *Service) Refresh(ctx context.Context, refresh string) (Session, error) {
	if refresh == "" {
		return Session{}, ErrUnauthorized
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Session{}, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	replacement := rand.Text() + rand.Text()
	row, err := apidb.New(tx).RotateRefreshToken(ctx, apidb.RotateRefreshTokenParams{OldHash: Hash(refresh), NewHash: Hash(replacement)})
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrUnauthorized
	}

	if err != nil {
		return Session{}, err
	}

	access, err := s.tokens.Issue(row.UserID, row.ID, time.Now())
	if err != nil {
		return Session{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Session{}, err
	}

	return Session{Access: access, Refresh: replacement, ExpiresAt: row.ExpiresAt}, nil
}

func (s *Service) Logout(ctx context.Context, access, refresh, expectedUser string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	q := apidb.New(tx)
	claims, parseErr := s.tokens.Parse(access, time.Now())
	if expectedUser != "" && parseErr == nil && claims.Subject != expectedUser {
		return ErrAccountChanged
	}

	if refresh != "" {
		owner, err := q.RefreshOwner(ctx, Hash(refresh))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if err == nil && expectedUser != "" && owner != expectedUser {
			return ErrAccountChanged
		}

		if err := q.RevokeRefreshToken(ctx, Hash(refresh)); err != nil {
			return err
		}
	}

	if parseErr == nil {
		if err := q.RevokeAuthSession(ctx, apidb.RevokeAuthSessionParams{ID: claims.SessionID, UserID: claims.Subject}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Service) UserID(ctx context.Context, access string) (string, error) {
	claims, err := s.tokens.Parse(access, time.Now())
	if err != nil {
		return "", err
	}

	id, err := apidb.New(s.db).GetAuthSessionUser(ctx, apidb.GetAuthSessionUserParams{ID: claims.SessionID, UserID: claims.Subject})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUnauthorized
	}

	return id, err
}

func (s *Service) Me(ctx context.Context, access string) (User, error) {
	claims, err := s.tokens.Parse(access, time.Now())
	if err != nil {
		return User{}, err
	}

	row, err := apidb.New(s.db).GetSessionProfile(ctx, apidb.GetSessionProfileParams{ID: claims.SessionID, UserID: claims.Subject})
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUnauthorized
	}

	if err != nil {
		return User{}, err
	}

	return User{ID: row.ID, Email: row.Email, Timezone: row.Timezone, CreatedAt: row.CreatedAt}, nil
}

func (s *Service) allowAttempt(ctx context.Context, email string) error {
	attempts, err := apidb.New(s.db).CountAuthAttempt(ctx, Hash(email))
	if err != nil {
		return err
	}

	if attempts > 10 {
		return ErrTooManyAttempts
	}

	return nil
}

func (s *Service) newSession(ctx context.Context, q *apidb.Queries, userID string) (Session, error) {
	now := time.Now().UTC()
	id := rand.Text()
	access, err := s.tokens.Issue(userID, id, now)
	if err != nil {
		return Session{}, err
	}

	session := Session{Access: access, Refresh: rand.Text() + rand.Text(), ExpiresAt: now.Add(RefreshTTL)}
	if err := q.CreateAuthSession(ctx, apidb.CreateAuthSessionParams{ID: id, UserID: userID, RefreshHash: Hash(session.Refresh), ExpiresAt: session.ExpiresAt}); err != nil {
		return Session{}, err
	}

	return session, nil
}
