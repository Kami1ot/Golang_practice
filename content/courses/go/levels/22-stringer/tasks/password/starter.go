package main

// LengthError — ошибка «пароль слишком короткий»:
// Got — фактическая длина, Min — минимально допустимая.
// Не меняйте объявление типа.
type LengthError struct {
	Got, Min int
}

// Error возвращает сообщение вида "длина пароля 3, нужно не меньше 8".
// Понадобится fmt.Sprintf — добавьте import "fmt".
func (e LengthError) Error() string {
	// TODO: соберите сообщение из полей e.Got и e.Min
	return ""
}

// CheckPassword возвращает nil, если пароль не короче 8 символов,
// и LengthError с заполненными полями — если короче.
// Длину считайте в символах (рунах), не в байтах!
func CheckPassword(password string) error {
	// TODO: сосчитайте руны и сравните с минимумом
	return nil
}
