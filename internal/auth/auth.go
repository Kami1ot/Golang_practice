// Package auth — криптопримитивы для паролей и токенов сессий.
// Только чистые функции: без HTTP и без SQL.
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Ограничения пароля. Верхняя граница — bcrypt молча обрезает всё после 72 байт.
const (
	MinPasswordBytes = 8
	MaxPasswordBytes = 72
)

func ValidatePassword(pw string) error {
	switch {
	case len(pw) < MinPasswordBytes:
		return fmt.Errorf("пароль короче %d символов", MinPasswordBytes)
	case len(pw) > MaxPasswordBytes:
		return fmt.Errorf("пароль длиннее %d байт", MaxPasswordBytes)
	}
	return nil
}

func HashPassword(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// NewToken — 32 байта crypto/rand в base64url (43 символа).
func NewToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
