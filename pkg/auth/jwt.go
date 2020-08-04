package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var secret string

func init() {
	secret = os.Getenv("JWT_SECRET")
	if len(secret) == 0 {
		panic("please provide JWT_SECRET")
	}
}

type Claims struct {
	jwt.StandardClaims
	Read  []string `json:"read"`
	Write []string `json:"write"`
}

func NewClaims(id string, read, write []string) Claims {
	return Claims{StandardClaims: jwt.StandardClaims{Subject: id, ExpiresAt: 0}, Read: read, Write: write}
}

func CreateToken(id string, read, write []string) (string, error) {
	claims := NewClaims(id, read, write)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	return tokenStr, err
}

func GetTokenClaims(tokenString string) (*Claims, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, errors.New("failed parsing token with claims")
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}

func WithClaims(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if len(tokenStr) > 0 {
			if !strings.Contains(tokenStr, " ") {
				http.Error(w, "no authorization header", http.StatusBadRequest)
				return
			}
			parts := strings.Split(tokenStr, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header. should be in form of: 'Bearer token'", http.StatusBadRequest)
				return
			}
			tokenStr = parts[1]
		} else {
			tokenStr = r.URL.Query().Get("token")
		}
		claims, err := GetTokenClaims(tokenStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "claims", claims)
		f(w, r.WithContext(ctx))
	}
}
