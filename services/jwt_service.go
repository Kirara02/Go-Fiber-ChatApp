package services

import (
	"errors"
	"main/config"
	"main/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
    GenerateTokens(user *domain.User) (accessToken string, refreshToken string, err error)
    GenerateAccessToken(user *domain.User) (string, error)
    ValidateRefreshToken(tokenString string) (*jwt.Token, error)
}

type jwtService struct {
    cfg *config.Config
}

func NewJWTService(cfg *config.Config) JWTService {
    return &jwtService{cfg: cfg}
}

func (s *jwtService) GenerateTokens(user *domain.User) (accessToken string, refreshToken string, err error) {
	
	accessToken, err = s.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshExp := time.Now().Add(time.Hour * 24 * time.Duration(s.cfg.RefreshTokenExpDays)).Unix()
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     refreshExp,
		"type":    "refresh",
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *jwtService) GenerateAccessToken(user *domain.User) (string, error) {
	accessExp := time.Now().Add(time.Hour * 24 * time.Duration(s.cfg.AccessTokenExpDays)).Unix()

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"name":    user.Name,
		"exp":     accessExp,
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (s *jwtService) ValidateRefreshToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["type"] != "refresh" {
			return nil, errors.New("invalid token type")
		}

		return []byte(s.cfg.JWTSecret), nil
	})
}