package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestCategorizeNil(t *testing.T) {
	if got := Categorize(nil); got != "ок" {
		t.Errorf("Categorize(nil) = %q, ожидалось %q", got, "ок")
	}
}

func TestCategorizeNotFoundDirect(t *testing.T) {
	if got := Categorize(ErrNotFound); got != "не найдено" {
		t.Errorf("Categorize(ErrNotFound) = %q, ожидалось %q", got, "не найдено")
	}
}

func TestCategorizeNotFoundWrapped(t *testing.T) {
	err := fmt.Errorf("обработка запроса: %w", fmt.Errorf("поиск пользователя: %w", ErrNotFound))
	if got := Categorize(err); got != "не найдено" {
		t.Errorf("Categorize с ErrNotFound под двумя обёртками = %q, ожидалось %q — используйте errors.Is", got, "не найдено")
	}
}

func TestCategorizeValidationWrapped(t *testing.T) {
	err := fmt.Errorf("сохранение формы: %w", ValidationError{Field: "email"})
	if got := Categorize(err); got != "невалидное поле: email" {
		t.Errorf("Categorize с обёрнутой ValidationError = %q, ожидалось %q — используйте errors.As", got, "невалидное поле: email")
	}
	deep := fmt.Errorf("api: %w", fmt.Errorf("вход: %w", ValidationError{Field: "age"}))
	if got := Categorize(deep); got != "невалидное поле: age" {
		t.Errorf("Categorize с ValidationError под двумя обёртками = %q, ожидалось %q", got, "невалидное поле: age")
	}
}

func TestCategorizeUnknown(t *testing.T) {
	if got := Categorize(errors.New("сеть отвалилась")); got != "неизвестная ошибка" {
		t.Errorf("Categorize(незнакомая ошибка) = %q, ожидалось %q", got, "неизвестная ошибка")
	}
}
