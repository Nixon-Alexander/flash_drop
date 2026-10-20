package helper

import (
	b64 "encoding/base64"
	"errors"
	"flash_drop/types"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

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

func CreateToken(username string, userId int, role types.UserRole) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"name":    username,
			"user_id": userId,
			"role":    role,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		},
	)

	secretKey := getSecretKey()
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (error, jwt.Claims) {
	if tokenString == "" {
		return errors.New("Token not found"), nil
	}

	claims := jwt.MapClaims{}
	secretKey := getSecretKey()
	parsedAccessToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return err, nil
	}

	if !parsedAccessToken.Valid {
		return errors.New("Token is invalid"), nil
	}

	return nil, parsedAccessToken.Claims.(jwt.MapClaims)
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
