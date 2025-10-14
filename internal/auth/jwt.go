package auth

import (
	"time"

	"datacenter-calc/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

type Claims struct {
	UserID uint           `json:"user_id"`
	Role   model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func InitJWT(secret string) {
	jwtSecret = []byte(secret)
}

func GenerateToken(userID uint, role model.UserRole, expMinutes int) (string, error) {
	exp := time.Now().Add(time.Duration(expMinutes) * time.Minute)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
