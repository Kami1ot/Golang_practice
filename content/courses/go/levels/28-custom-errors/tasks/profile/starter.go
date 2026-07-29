package main

import "errors"

// Sentinel-ошибки сервиса. Не меняйте.
var (
	ErrEmptyName = errors.New("имя не задано")
	ErrBanned    = errors.New("пользователь заблокирован")
)

// NotFoundError — ошибка «профиль не найден» с именем внутри. Не меняйте объявление.
type NotFoundError struct {
	Name string
}

// Error возвращает сообщение вида: профиль "vasya" не найден
// Понадобится fmt (глагол %q сам добавляет кавычки).
func (e NotFoundError) Error() string {
	// TODO
	return ""
}

// OpenProfile проверяет доступ к профилю:
// пустое имя → ErrEmptyName; нет в карте → NotFoundError;
// статус "banned" → обёрнутый ErrBanned ("доступ к <имя>: ..."); иначе nil.
func OpenProfile(statuses map[string]string, name string) error {
	// TODO: три guard-проверки по порядку
	return nil
}
