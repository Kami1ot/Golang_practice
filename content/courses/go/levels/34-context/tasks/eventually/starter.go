package main

import "time"

// Eventually ждёт значение из result, но не дольше timeout.
// Успели — значение и nil; нет — "" и ошибка контекста.
// Понадобится context.WithTimeout — добавьте import "context".
func Eventually(result <-chan string, timeout time.Duration) (string, error) {
	// TODO: контекст со сроком + select на два случая
	return "", nil
}
