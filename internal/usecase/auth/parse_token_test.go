package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestParseToken(t *testing.T) {
	secretKey := "secret"
	secretBytes := []byte(secretKey)

	tests := []struct {
		name        string
		tokenString string
		wantUserID  int
		wantErr     bool
	}{
		{
			name: "Success parsing valid token",
			tokenString: createTestToken(t, secretBytes,
				123,
				jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			),
			wantUserID: 123,
			wantErr:    false,
		},
		{
			name:        "Error parsing invalid token",
			tokenString: "invalid-token",
			wantUserID:  0,
			wantErr:     true,
		},
		{
			name: "Error parsing expired token",
			tokenString: createTestToken(t,
				secretBytes,
				123,
				jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			),
			wantUserID: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewAuthUsecase(nil, nil, secretKey, time.Hour)

			claims, err := uc.ParseToken(context.Background(), tt.tokenString)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantUserID, claims.UserID)
		})
	}
}

func createTestToken(t *testing.T,
	secretKey []byte,
	userID int,
	regClaims jwt.RegisteredClaims,
) string {
	tokenClaims := tokenClaims{
		RegisteredClaims: regClaims,
		UserID:           userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, err := token.SignedString(secretKey)
	assert.NoError(t, err)

	return tokenString
}
