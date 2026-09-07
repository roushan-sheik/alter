package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	jwtAccessSecretKey  = "your_access_secret_key"
	jwtRefreshSecretKey = "your_refresh_secret_key"
)

type JwtCustomClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsPremium bool      `json:"is_premium"`
	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateToken(userId uuid.UUID, name string, email string, role string, isPremium bool) (string, string, error)
	GenerateResetToken(userId uuid.UUID, email string) (string, error)
	ValidateToken(tokenString string, isRefresh bool) (*JwtCustomClaims, error)
	ValidateResetToken(tokenString string) (*JwtCustomClaims, error)
}

type jwtService struct {
	jwtAccessSecretKey  []byte
	jwtRefreshSecretKey []byte
	accessTokenExp      time.Duration
	refreshTokenExp     time.Duration
}

func NewJWTService(accessSecretKey, refreshSecretKey, accessExpiry, refreshExpiry string) JWTService {

	if accessSecretKey == "" {
		accessSecretKey = jwtAccessSecretKey
	}
	if refreshSecretKey == "" {
		refreshSecretKey = jwtRefreshSecretKey
	}

	accExp, err := time.ParseDuration(accessExpiry)
	if err != nil || accExp == 0 {
		accExp = 24 * time.Hour
	}

	refExp, err := time.ParseDuration(refreshExpiry)
	if err != nil || refExp == 0 {
		refExp = 30 * 24 * time.Hour
	}

	return &jwtService{
		jwtAccessSecretKey:  []byte(accessSecretKey),
		jwtRefreshSecretKey: []byte(refreshSecretKey),
		accessTokenExp:      accExp,
		refreshTokenExp:     refExp,
	}
}

func (js *jwtService) GenerateToken(userId uuid.UUID, name string, email string, role string, isPremium bool) (string, string, error) {
	accessClaims := &JwtCustomClaims{
		UserID:    userId,
		Name:      name,
		Email:     email,
		Role:      role,
		IsPremium: isPremium,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(js.accessTokenExp)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	aToken, err := accessToken.SignedString(js.jwtAccessSecretKey)
	if err != nil {
		return "", "", err
	}

	refreshClaims := &JwtCustomClaims{
		UserID:    userId,
		Name:      name,
		Email:     email,
		Role:      role,
		IsPremium: isPremium,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(js.refreshTokenExp)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rToken, err := refreshToken.SignedString(js.jwtRefreshSecretKey)
	if err != nil {
		return "", "", err
	}

	return aToken, rToken, nil
}

func (js *jwtService) GenerateResetToken(userId uuid.UUID, email string) (string, error) {
	resetClaims := &JwtCustomClaims{
		UserID: userId,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Subject:   "reset",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, resetClaims)
	return token.SignedString(js.jwtAccessSecretKey)
}

func (js *jwtService) ValidateToken(tokenString string, isRefresh bool) (*JwtCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if isRefresh {
			return js.jwtRefreshSecretKey, nil
		}
		return js.jwtAccessSecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JwtCustomClaims); ok && token.Valid {
		if claims.Subject == "reset" {
			return nil, errors.New("invalid token: expected access token")
		}
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (js *jwtService) ValidateResetToken(tokenString string) (*JwtCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return js.jwtAccessSecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JwtCustomClaims); ok && token.Valid {
		if claims.Subject != "reset" {
			return nil, errors.New("invalid token: expected reset token")
		}
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
