package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestJWTManager_GenerateAndParse(t *testing.T) {
	m := NewJWTManager("access-secret", "refresh-secret", 15, 24)
	userID := uuid.New()

	pair, err := m.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("tokens should not be empty")
	}

	accessClaims, err := m.ParseAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if accessClaims.UserID != userID.String() {
		t.Fatalf("unexpected user id in access token")
	}

	refreshClaims, err := m.ParseRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if refreshClaims.UserID != userID.String() {
		t.Fatalf("unexpected user id in refresh token")
	}
}
