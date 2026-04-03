package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RoleClaim struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type AccessClaims struct {
	Roles       []RoleClaim `json:"roles"`
	Permissions []string    `json:"permissions"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	jwt.RegisteredClaims
}

func GenerateAccessToken(secret string, userID string, deviceID string, roles []RoleClaim, permissions []string, expiry time.Duration) (string, error) {
	claims := AccessClaims{
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Audience:  jwt.ClaimStrings{deviceID},
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(secret string, userID string, deviceID string, expiry time.Duration) (string, error) {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Audience:  jwt.ClaimStrings{deviceID},
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseAccessToken(secret string, tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func ParseRefreshToken(secret string, tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

// AllPermissions merges unique permissions from roles + direct permissions
func (c *AccessClaims) AllPermissions() []string {
	seen := make(map[string]bool)
	var perms []string
	for _, role := range c.Roles {
		for _, p := range role.Permissions {
			if !seen[p] {
				seen[p] = true
				perms = append(perms, p)
			}
		}
	}
	for _, p := range c.Permissions {
		if !seen[p] {
			seen[p] = true
			perms = append(perms, p)
		}
	}
	return perms
}

// GetUserID returns sub claim
func (c *AccessClaims) GetUserID() string {
	sub, _ := c.GetSubject()
	return sub
}

// GetDeviceID returns first aud claim
func (c *AccessClaims) GetDeviceID() string {
	aud, _ := c.GetAudience()
	if len(aud) > 0 {
		return aud[0]
	}
	return "na"
}

// GetUserID returns sub claim
func (c *RefreshClaims) GetUserID() string {
	sub, _ := c.GetSubject()
	return sub
}

// GetDeviceID returns first aud claim
func (c *RefreshClaims) GetDeviceID() string {
	aud, _ := c.GetAudience()
	if len(aud) > 0 {
		return aud[0]
	}
	return "na"
}
