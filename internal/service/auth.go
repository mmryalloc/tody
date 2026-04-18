package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mmryalloc/todo-app/internal/auth"
	"github.com/mmryalloc/todo-app/internal/entity"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSession     = errors.New("invalid or expired session")
)

type UserRepository interface {
	Create(ctx context.Context, u *entity.User) error
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	GetByID(ctx context.Context, id int64) (entity.User, error)
}

type SessionRepository interface {
	Save(ctx context.Context, userID int64, tokenHash string, s entity.Session, ttl time.Duration) error
	Exists(ctx context.Context, userID int64, tokenHash string) (bool, error)
	LookupUserID(ctx context.Context, tokenHash string) (int64, error)
	Delete(ctx context.Context, userID int64, tokenHash string) error
	DeleteAllForUser(ctx context.Context, userID int64) error
}

type TokenIssuer interface {
	Generate(userID int64) (string, time.Time, error)
}

type SessionContext struct {
	UserAgent string
	IPAddress string
}

type TokenPair struct {
	UserID       int64
	AccessToken  string
	RefreshToken string
}

type authService struct {
	users           UserRepository
	sessions        SessionRepository
	tokens          TokenIssuer
	refreshTokenTTL time.Duration
}

func NewAuthService(users UserRepository, sessions SessionRepository, tokens TokenIssuer, refreshTokenTTL time.Duration) *authService {
	return &authService{
		users:           users,
		sessions:        sessions,
		tokens:          tokens,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *authService) Register(ctx context.Context, email, password string) (entity.User, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return entity.User{}, fmt.Errorf("service auth register: %w", err)
	}

	u := entity.User{
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: hash,
	}

	if err := s.users.Create(ctx, &u); err != nil {
		if errors.Is(err, entity.ErrUserExists) {
			return entity.User{}, ErrEmailTaken
		}
		return entity.User{}, fmt.Errorf("service auth register: %w", err)
	}

	return u, nil
}

func (s *authService) Login(ctx context.Context, email, password string, sc SessionContext) (TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, fmt.Errorf("service auth login lookup: %w", err)
	}

	if err := auth.VerifyPassword(u.PasswordHash, password); err != nil {
		if errors.Is(err, auth.ErrInvalidPassword) {
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, fmt.Errorf("service auth login verify: %w", err)
	}

	return s.issueTokenPair(ctx, u.ID, sc)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string, sc SessionContext) (TokenPair, error) {
	oldHash := auth.HashRefreshToken(refreshToken)

	userID, err := s.sessions.LookupUserID(ctx, oldHash)
	if err != nil {
		if errors.Is(err, entity.ErrSessionNotFound) {
			return TokenPair{}, ErrInvalidSession
		}
		return TokenPair{}, fmt.Errorf("service auth refresh lookup: %w", err)
	}

	pair, err := s.issueTokenPair(ctx, userID, sc)
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.sessions.Delete(ctx, userID, oldHash); err != nil {
		return TokenPair{}, fmt.Errorf("service auth refresh revoke old: %w", err)
	}

	return pair, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	hash := auth.HashRefreshToken(refreshToken)

	userID, err := s.sessions.LookupUserID(ctx, hash)
	if err != nil {
		if errors.Is(err, entity.ErrSessionNotFound) {
			return nil
		}
		return fmt.Errorf("service auth logout lookup: %w", err)
	}

	if err := s.sessions.Delete(ctx, userID, hash); err != nil {
		return fmt.Errorf("service auth logout: %w", err)
	}
	return nil
}

func (s *authService) LogoutAll(ctx context.Context, userID int64) error {
	if err := s.sessions.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("service auth logout all: %w", err)
	}
	return nil
}

func (s *authService) issueTokenPair(ctx context.Context, userID int64, sc SessionContext) (TokenPair, error) {
	accessToken, _, err := s.tokens.Generate(userID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("service auth issue access: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("service auth issue refresh: %w", err)
	}

	session := entity.Session{
		UserAgent: sc.UserAgent,
		IPAddress: sc.IPAddress,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.sessions.Save(ctx, userID, auth.HashRefreshToken(refreshToken), session, s.refreshTokenTTL); err != nil {
		return TokenPair{}, fmt.Errorf("service auth issue persist: %w", err)
	}

	return TokenPair{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
