package helper

import (
	"fiber/skp/app/model"
	"fiber/skp/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(u model.User, permissions []string) (string, error) {
	claims := model.JWTClaims{
		UserID:      u.ID,
		Username:    u.Username,
		Role:        u.Role.Name,
		Permissions: permissions,
		Type:        "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	secret := config.GetJWTSecret()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(u model.User) (string, error) {
	claims := model.JWTClaims{
		UserID:   u.ID,
		Username: u.Username,
		Role:     u.Role.Name,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	secret := config.GetJWTSecret()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string) (*model.JWTClaims, error) {
	secret := config.GetJWTSecret()
	token, err := jwt.ParseWithClaims(tokenString, &model.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*model.JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
