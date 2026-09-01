package service

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/config"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters long")
	ErrNameRequired       = errors.New("name is required")
	ErrEmailRequired      = errors.New("email is required")
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type AuthService interface {
	Register(name, email, password string) (*model.User, string, string, error)
	Login(email, password string) (*model.User, string, string, error)
	RefreshToken(refreshTokenString string) (string, string, error)
	GetUserByID(id uuid.UUID) (*model.User, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.JWTConfig
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.JWTConfig) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *authService) Register(name, email, password string) (*model.User, string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "", "", ErrNameRequired
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, "", "", ErrEmailRequired
	}

	if len(password) < 6 {
		return nil, "", "", ErrPasswordTooShort
	}

	existing, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, "", "", err
	}
	if existing != nil {
		return nil, "", "", ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	user := &model.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", err
	}

	accessToken, refreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Login(email, password string) (*model.User, string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, "", "", err
	}
	if user == nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(refreshTokenString string) (string, string, error) {
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil || user == nil {
		return "", "", ErrInvalidToken
	}

	return s.generateTokenPair(user)
}

func (s *authService) GetUserByID(id uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *authService) generateTokenPair(user *model.User) (string, string, error) {
	accessExpiry := time.Duration(s.cfg.AccessExpiryMinutes) * time.Minute
	if accessExpiry == 0 {
		accessExpiry = 15 * time.Minute
	}

	refreshExpiry := time.Duration(s.cfg.RefreshExpiryDays) * 24 * time.Hour
	if refreshExpiry == 0 {
		refreshExpiry = 7 * 24 * time.Hour
	}

	now := time.Now()

	// Access Token
	accessClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
			Issuer:    "taskflow-api",
		},
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessTokenObj.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
			Issuer:    "taskflow-api",
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
