package services

import (
	"errors"
	"fmt"
	"main/domain"
	"main/dto"
	"main/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(req *dto.RegisterRequest) (*domain.User, error)
	Login(req *dto.LoginRequest) (accessToken string, refreshToken string, user *domain.User, err error)
	RefreshToken(req *dto.RefreshTokenRequest) (newAccessToken string, newRefreshToken string, err error)
	Logout(tokenString string) error
}

type authService struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.TokenRepository
	jwtService JWTService
}

func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, jwtService JWTService) AuthService {
	return &authService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtService: jwtService,
	}
}


func (s *authService) Register(req *dto.RegisterRequest) (*domain.User, error) {
	_, err := s.userRepo.GetUserByEmail(req.Email)
	if err == nil {
		return nil, errors.New("email already registered")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(req *dto.LoginRequest) (string, string, *domain.User, error) {
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return "", "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", "", nil, err
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokens(user)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}


func (s *authService) RefreshToken(req *dto.RefreshTokenRequest) (string, string, error) {
	token, err := s.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil || !token.Valid {
		return "", "", errors.New("refresh token tidak valid atau kedaluwarsa")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("gagal mendapatkan klaim token")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return "", "", errors.New("ID pengguna tidak valid di dalam token")
	}
	userID := uint(userIDFloat)

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return "", "", errors.New("pengguna tidak ditemukan")
	}

	newAccessToken, newRefreshToken, err := s.jwtService.GenerateTokens(user)
	if err != nil {
		return "", "", fmt.Errorf("gagal membuat token baru: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) Logout(tokenString string) error {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return fmt.Errorf("gagal parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("gagal mendapatkan klaim dari token")
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("klaim 'exp' tidak valid")
	}
	expiresAt := time.Unix(int64(expFloat), 0)

	invalidatedToken := &domain.InvalidatedToken{
		Token:     tokenString,
		ExpiresAt: expiresAt,
	}

	if err := s.tokenRepo.CreateInvalidatedToken(invalidatedToken); err != nil {
		return fmt.Errorf("gagal menyimpan token ke denylist: %w", err)
	}

	return nil
}