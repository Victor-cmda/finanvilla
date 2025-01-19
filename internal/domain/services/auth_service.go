package services

import (
	"context"
	"finanvilla/internal/application/dtos"
	"finanvilla/internal/domain/entities"
	"finanvilla/internal/domain/enums"
	"finanvilla/internal/domain/repositories"
	"finanvilla/pkg/errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type AuthService struct {
	userService        *UserService
	refreshTokenRepo   repositories.RefreshTokenRepository
	jwtSecret          string
	refreshTokenSecret string
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
}

func NewAuthService(
	userService *UserService,
	refreshTokenRepo repositories.RefreshTokenRepository,
	jwtSecret string,
	refreshTokenSecret string,
) *AuthService {
	return &AuthService{
		userService:        userService,
		refreshTokenRepo:   refreshTokenRepo,
		jwtSecret:          jwtSecret,
		refreshTokenSecret: refreshTokenSecret,
		accessTokenTTL:     15 * time.Minute,   // Token JWT expira em 15 minutos
		refreshTokenTTL:    7 * 24 * time.Hour, // Refresh token expira em 7 dias
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Register(ctx context.Context, req *dtos.RegisterRequest) (*entities.User, error) {
	exists, err := s.userService.GetByEmail(ctx, req.Email)
	if err != errors.ErrUserNotFound {
		return nil, err
	}

	if exists != nil {
		return nil, errors.ErrEmailAlreadyUsed
	}

	user := &entities.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		UserType:  enums.Standard,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.userService.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *dtos.LoginRequest) (*TokenPair, error) {
	user, err := s.userService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}
	if err := s.refreshTokenRepo.RevokeByUserID(ctx, userID); err != nil {
		return nil, err
	}

	return s.generateTokenPair(ctx, user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	rt, err := s.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if rt.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}

	user, err := s.userService.GetByID(ctx, rt.UserID.String())
	if err != nil {
		return nil, err
	}

	if err := s.refreshTokenRepo.RevokeToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	return s.generateTokenPair(ctx, user)
}

func (s *AuthService) generateTokenPair(ctx context.Context, user *entities.User) (*TokenPair, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}

	refreshToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
	}, nil
}

func (s *AuthService) generateAccessToken(user *entities.User) (string, error) {
	claims := jwt.MapClaims{
		"userId": user.ID,
		"email":  user.Email,
		"exp":    time.Now().Add(s.accessTokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return errors.ErrInternalServer
	}

	err := s.refreshTokenRepo.RevokeToken(ctx, refreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.ErrInternalServer
	}

	err := s.refreshTokenRepo.RevokeByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}
