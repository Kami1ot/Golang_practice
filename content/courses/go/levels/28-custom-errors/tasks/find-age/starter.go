package main

import "errors"

// ErrUserNotFound — сигнальная ошибка «пользователь не найден». Не меняйте.
var ErrUserNotFound = errors.New("пользователь не найден")

// FindAge возвращает возраст пользователя из карты.
// Если имени нет — 0 и ErrUserNotFound (именно это значение!).
func FindAge(users map[string]int, name string) (int, error) {
	// TODO: ok-идиома карты + возврат sentinel
	return 0, nil
}
