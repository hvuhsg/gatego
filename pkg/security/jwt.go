package security

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
)

// GenerateJWT generates a JWT token with the given user ID and expiration time.
func GenerateJWT(claims jwt.Claims, secretKey string) (string, error) {
	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with a secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates the JWT token and populate the claims if valid.
func ValidateJWT(tokenString string, secretKey string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		return err
	}

	// Check if the token is valid
	if !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}
