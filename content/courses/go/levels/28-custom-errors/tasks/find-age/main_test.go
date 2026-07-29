package main

import (
	"errors"
	"testing"
)

func TestFindAgeFound(t *testing.T) {
	users := map[string]int{"Аня": 25, "Боря": 30}
	got, err := FindAge(users, "Аня")
	if err != nil {
		t.Fatalf("FindAge(users, \"Аня\") вернула ошибку %v, ожидался nil", err)
	}
	if got != 25 {
		t.Errorf("FindAge(users, \"Аня\") = %d, ожидалось 25", got)
	}
}

func TestFindAgeNotFound(t *testing.T) {
	users := map[string]int{"Аня": 25}
	got, err := FindAge(users, "Юра")
	if err == nil {
		t.Fatalf("FindAge несуществующего имени вернула nil, ожидалась ошибка")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("errors.Is(err, ErrUserNotFound) = false — возвращайте сам sentinel, а не новый errors.New с тем же текстом")
	}
	if got != 0 {
		t.Errorf("при ошибке возраст = %d, ожидалось 0", got)
	}
}

func TestFindAgeSameSentinel(t *testing.T) {
	users := map[string]int{}
	_, err1 := FindAge(users, "а")
	_, err2 := FindAge(users, "б")
	if err1 != err2 {
		t.Errorf("два вызова вернули РАЗНЫЕ значения ошибки — sentinel должен быть одним значением, объявленным один раз")
	}
}

func TestFindAgeZeroAge(t *testing.T) {
	users := map[string]int{"малыш": 0}
	got, err := FindAge(users, "малыш")
	if err != nil {
		t.Fatalf("FindAge существующего имени с возрастом 0 вернула ошибку %v — ноль в карте это не «не найдено», используйте ok-идиому", err)
	}
	if got != 0 {
		t.Errorf("FindAge(users, \"малыш\") = %d, ожидалось 0", got)
	}
}
