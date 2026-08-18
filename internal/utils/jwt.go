package utils

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/config"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed JWT token for the given user.
func GenerateJWT(userID uint, phone, role string) (string, error) {
	cfg := config.AppConfig
	expiryHours, err := strconv.Atoi(cfg.JWTExpiryHours)
	if err != nil {
		expiryHours = 72
	}
	return GenerateJWTWithExpiry(userID, phone, role, time.Now().Add(time.Duration(expiryHours)*time.Hour))
}

// GenerateJWTWithExpiry creates a signed JWT with an explicit expiry
// instant. GenerateJWT is the normal entry point for login flows; this
// exists so tests (e.g. role/middleware tests) can deterministically
// construct an already-expired token without sleeping.
func GenerateJWTWithExpiry(userID uint, phone, role string, expiresAt time.Time) (string, error) {
	cfg := config.AppConfig

	claims := Claims{
		UserID: userID,
		Phone:  phone,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ValidateJWT parses and validates a JWT token string, returning its claims.
func ValidateJWT(tokenString string) (*Claims, error) {
	cfg := config.AppConfig
	claims := &Claims{}

	// WithValidMethods pins parsing to HS256 — without it, a token crafted
	// with a different algorithm (e.g. "none") could bypass verification,
	// since the keyfunc alone doesn't constrain which algorithm is accepted.
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
