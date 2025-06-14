package token

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type CustomClaims struct {
	Email string `json:"email"`
	Role  Role   `json:"role"`
	jwt.RegisteredClaims
}

type TokenValidationResult struct {
	Valid    bool
	Expired  bool
	Claims   *CustomClaims
	ParseErr error
}

type TokenManager interface {
	Generate(email string, role Role) (string, error)
	Validate(tokenStr string) *TokenValidationResult
}

type JWTManager struct {
	secretKey []byte
	duration  time.Duration
}

func NewJWTManager(secret string, duration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secret),
		duration:  duration,
	}
}

func (j *JWTManager) Generate(email string, role Role) (string, error) {
	if !isValidRole(role) {
		return "", fmt.Errorf("Invalid Role: %s", role)
	}

	claims := &CustomClaims{
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.duration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWTManager) Validate(tokenStr string) *TokenValidationResult {
	result := &TokenValidationResult{}

	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})

	if err != nil {
		result.ParseErr = err
	}

	if token == nil {
		return result
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		result.Claims = claims
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			result.Expired = true
		} else if token.Valid {
			result.Valid = true
		}
	}

	return result
}
