package helper

import (
	b64 "encoding/base64"
	"errors"
	"flash_drop/types"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthClaims struct {
	UserID int            `json:"user_id"`
	Name   string         `json:"name"`
	Role   types.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte { return []byte(os.Getenv("JWT_SECRET")) }

func getSecretKey() []byte {
	return []byte("secret_key")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateToken(name string, userID int, role types.UserRole) (string, error) {
	now := time.Now()
	claims := AuthClaims{
		UserID: userID,
		Name:   name,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "flash_drop",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
}

func VerifyToken(tokenString string) (*AuthClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token not found")
	}
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) { return jwtSecret(), nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer("flash_drop"),
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	return claims, nil
}

func MakeSlug(s string) string {
	// Convert to lowercase
	result := strings.ToLower(s)

	// Replace spaces with hyphens
	result = strings.ReplaceAll(result, " ", "-")

	// Remove non-alphanumeric characters except for hyphens
	reg, err := regexp.Compile("[^a-z0-9-]+")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	result = reg.ReplaceAllString(result, "")

	// Return the slug
	return result
}

func EncodeBase64(s string) string {
	data := fmt.Sprint(s)
	sEnc := b64.StdEncoding.EncodeToString([]byte(data))
	return sEnc
}
