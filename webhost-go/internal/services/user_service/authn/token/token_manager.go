package token

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type CustomClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type TokenValidationResult struct {
	Valid    bool
	Expired  bool
	Claims   *CustomClaims
	ParseErr error
}

type TokenManager interface {
	Generate(email string, role string) (string, error)
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

func (j *JWTManager) Generate(email string, role string) (string, error) {
	claims := &CustomClaims{
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * j.duration)),
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
