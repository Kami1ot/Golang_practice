package auth

import (
	"strings"
	"testing"
)

func TestPasswordRoundtrip(t *testing.T) {
	hash, err := HashPassword("секретный-пароль")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "секретный-пароль") {
		t.Fatal("верный пароль не прошёл проверку")
	}
	if CheckPassword(hash, "неверный") {
		t.Fatal("неверный пароль прошёл проверку")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("1234567"); err == nil {
		t.Fatal("короткий пароль должен отклоняться")
	}
	if err := ValidatePassword(strings.Repeat("a", 73)); err == nil {
		t.Fatal("пароль длиннее 72 байт должен отклоняться (bcrypt обрезает)")
	}
	if err := ValidatePassword("нормальный-пароль"); err != nil {
		t.Fatalf("валидный пароль отклонён: %v", err)
	}
}

func TestNewTokenUnique(t *testing.T) {
	a, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := NewToken()
	if a == b || len(a) < 40 {
		t.Fatalf("токены подозрительные: %q %q", a, b)
	}
}
