package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing a password: %w", err)
	}

	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("hash and password do not match: %w", err)
	}

	return nil
}

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userId.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}

		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse or validate token: %w", err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("token is invalid: %w", err)
	}

	userIdRaw, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get subject from token claims: %w", err)
	}

	if userIdRaw == "" {
		return uuid.Nil, fmt.Errorf("token subject is missing: %w", err)
	}

	userId, err := uuid.Parse(userIdRaw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not parse string to uuid: %w", err)
	}

	return userId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("authorization header empty")
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("expected authorization header split to have two parts")
	}

	if strings.TrimSpace(parts[0]) != "Bearer" {
		return "", fmt.Errorf("expected Bearer, instead got %s", parts[0])
	}

	token := strings.TrimSpace(parts[1])

	return token, nil
}

func MakeRefreshToken() (string, error) {
	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("error making refresh token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("authorization header empty")
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("expected authorization header split to have two parts")
	}

	if strings.TrimSpace(parts[0]) != "ApiKey" {
		return "", fmt.Errorf("expected Bearer, instead got %s", parts[0])
	}

	token := strings.TrimSpace(parts[1])
	return token, nil
}
