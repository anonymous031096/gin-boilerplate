package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewJWTManager(accessSecret, refreshSecret string, accessTTLMin, refreshTTLHour int) *JWTManager {
	return &JWTManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     time.Duration(accessTTLMin) * time.Minute,
		refreshTTL:    time.Duration(refreshTTLHour) * time.Hour,
	}
}

func (m *JWTManager) GenerateTokenPair(userID uuid.UUID) (TokenPair, error) {
	access, err := m.sign(userID, "access", m.accessSecret, m.accessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := m.sign(userID, "refresh", m.refreshSecret, m.refreshTTL)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (m *JWTManager) ParseAccessToken(token string) (*Claims, error) {
	return m.parse(token, m.accessSecret, "access")
}

func (m *JWTManager) ParseRefreshToken(token string) (*Claims, error) {
	return m.parse(token, m.refreshSecret, "refresh")
}

func (m *JWTManager) sign(userID uuid.UUID, tokenType string, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID.String(),
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

func (m *JWTManager) parse(token string, secret []byte, expectedType string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Type != expectedType {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}
