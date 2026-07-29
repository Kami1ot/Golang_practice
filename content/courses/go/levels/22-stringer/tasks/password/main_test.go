package main

import "testing"

// Компилируется только если LengthError реализует error.
var _ error = LengthError{}

func TestErrorMessage(t *testing.T) {
	e := LengthError{Got: 3, Min: 8}
	want := "длина пароля 3, нужно не меньше 8"
	if got := e.Error(); got != want {
		t.Errorf("LengthError{3, 8}.Error() = %q, ожидалось %q", got, want)
	}
}

func TestShortAscii(t *testing.T) {
	err := CheckPassword("abc")
	if err == nil {
		t.Fatalf("CheckPassword(\"abc\") = nil, ожидалась ошибка: 3 символа из 8")
	}
	le, ok := err.(LengthError)
	if !ok {
		t.Fatalf("CheckPassword должен возвращать именно LengthError, получено %T", err)
	}
	if le.Got != 3 || le.Min != 8 {
		t.Errorf("поля ошибки = {Got: %d, Min: %d}, ожидалось {Got: 3, Min: 8}", le.Got, le.Min)
	}
}

func TestShortCyrillic(t *testing.T) {
	err := CheckPassword("пароль")
	if err == nil {
		t.Fatalf("CheckPassword(\"пароль\") = nil, ожидалась ошибка: в слове 6 символов")
	}
	le, ok := err.(LengthError)
	if !ok {
		t.Fatalf("CheckPassword должен возвращать именно LengthError, получено %T", err)
	}
	if le.Got != 6 {
		t.Errorf("Got = %d, ожидалось 6 — длину считаем в рунах, а не в байтах (len даст 12)", le.Got)
	}
}

func TestValidPasswords(t *testing.T) {
	for _, p := range []string{"12345678", "go-это-сила!", "длинный-пароль"} {
		if err := CheckPassword(p); err != nil {
			t.Errorf("CheckPassword(%q) = %v, ожидался nil — пароль достаточно длинный", p, err)
		}
	}
}

func TestBoundary(t *testing.T) {
	if err := CheckPassword("абвгдежз"); err != nil {
		t.Errorf("CheckPassword на пароле ровно из 8 символов = %v, ожидался nil — граница входит в норму", err)
	}
	if err := CheckPassword("абвгдеж"); err == nil {
		t.Errorf("CheckPassword на пароле из 7 символов = nil, ожидалась ошибка")
	}
}
