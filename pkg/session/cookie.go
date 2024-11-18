package session

import (
	"maps"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hvuhsg/gatego/pkg/security"
)

type JWTCookie struct {
	*http.Cookie
}

func (jc *JWTCookie) GetItems(key string) (map[string]any, error) {
	claims := &jwt.MapClaims{}
	err := security.ValidateJWT(jc.Value, key, claims)
	if err != nil {
		return nil, err
	}

	return *claims, nil
}

func (jc *JWTCookie) SetItems(key string, items map[string]any) error {
	claims := jwt.MapClaims{}
	maps.Copy(claims, items)
	value, err := security.GenerateJWT(claims, key)
	if err != nil {
		return err
	}

	jc.Value = value

	return nil
}
