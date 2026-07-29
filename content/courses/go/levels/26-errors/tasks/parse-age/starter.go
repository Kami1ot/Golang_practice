package main

// ParseAge разбирает возраст из строки анкеты:
// не число — пробросить ошибку strconv.Atoi;
// вне 0–150 — своя ошибка "возраст N вне диапазона [0, 150]";
// иначе — возраст и nil.
func ParseAge(s string) (int, error) {
	// TODO: strconv.Atoi → проверка err → проверка диапазона → счастливый путь
	return 0, nil
}
