package services

import (
	"context"
	"finanvilla/internal/application/dtos"
	"finanvilla/internal/domain/entities"
	"finanvilla/internal/domain/enums"
	"finanvilla/pkg/errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AuthService struct {
	userService *UserService
	jwtSecret   []byte
}

func NewAuthService(userService *UserService, jwtSecret string) *AuthService {
	return &AuthService{
		userService: userService,
		jwtSecret:   []byte(jwtSecret),
	}
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

func (s *AuthService) Login(ctx context.Context, req *dtos.LoginRequest) (string, error) {
	user, err := s.userService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"user_type": user.UserType,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
