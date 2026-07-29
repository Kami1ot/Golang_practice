package main

import (
	"errors"
	"fmt"
)

// ErrNotFound — известная ошибка «не найдено». Не меняйте.
var ErrNotFound = errors.New("не найдено")

// ValidationError — ошибка валидации с именем поля. Не меняйте.
type ValidationError struct {
	Field string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("поле %s не прошло валидацию", e.Field)
}

// Categorize раскладывает ошибку по категориям:
// nil → "ок"; ErrNotFound в цепочке → "не найдено";
// ValidationError в цепочке → "невалидное поле: <Field>";
// иначе → "неизвестная ошибка".
func Categorize(err error) string {
	// TODO: nil → errors.Is → errors.As → общий случай
	return ""
}
