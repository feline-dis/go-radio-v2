package services

import (
	"errors"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret     []byte
	expiration time.Duration
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		secret:     []byte(cfg.JWT.Secret),
		expiration: cfg.JWT.Expiration,
	}
}

// GenerateToken creates a new JWT token for the given username
func (j *JWTService) GenerateToken(username string) (string, error) {
	if len(j.secret) == 0 {
		return "", errors.New("JWT secret not configured")
	}

	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken validates and parses a JWT token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	if len(j.secret) == 0 {
		return nil, errors.New("JWT secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken creates a new token with extended expiration for valid tokens
func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate a new token with extended expiration
	return j.GenerateToken(claims.Username)
} 