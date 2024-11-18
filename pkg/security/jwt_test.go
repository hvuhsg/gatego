package security

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestGenerateAndValidate(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name      string
		claims    jwt.StandardClaims
		secretKey string
	}{
		{
			name: "ValidToken",
			claims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				Audience:  "aud",
			},
			secretKey: "secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate JWT token
			tokenString, err := GenerateJWT(tt.claims, tt.secretKey)

			// Check error
			if err != nil {
				t.Errorf("GenerateJWT() error = %v", err)
				return
			}

			// Validate token format
			if tokenString == "" {
				t.Errorf("GenerateJWT() returned an empty token string")
				return
			}

			claims := &jwt.StandardClaims{}
			err = ValidateJWT(tokenString, tt.secretKey, claims)
			if err != nil {
				t.Errorf("ValidateJWT() error = %v", err)
				return
			}

			if claims.Audience != tt.claims.Audience {
				t.Errorf("ValidateJWT() returned audience %s and expected audience %s", claims.Audience, tt.claims.Audience)
				return
			}
		})
	}
}
