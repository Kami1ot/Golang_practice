package main

import "errors"

// ErrEmpty — известная ошибка пустого ввода. Не меняйте.
var ErrEmpty = errors.New("пустой ввод")

// Validate возвращает ErrEmpty для пустой строки. Не меняйте.
func Validate(s string) error {
	if s == "" {
		return ErrEmpty
	}
	return nil
}

// Process вызывает Validate и оборачивает её ошибку с контекстом "обработка: ".
// Исходная ошибка должна остаться находимой через errors.Is.
func Process(s string) error {
	// TODO: вызов Validate + обёртка через fmt.Errorf (import "fmt")
	return nil
}
