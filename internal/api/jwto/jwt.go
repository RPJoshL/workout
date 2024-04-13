package jwto

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserId int `json:"user_id"`
	jwt.RegisteredClaims
}

// ValidateToken validates the given token. Authroized returns if the token and
// the expiry date were still valid
func ValidateToken(token string, key []byte) (claim *Claims, authorized bool, err error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil || !tkn.Valid {
		return nil, false, err
	}

	return claims, true, nil
}

// CreateToken Creates a new JWT token and returns it
func CreateToken(key []byte, claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign the JWT token: %s", err)
	}

	return tokenStr, nil
}
