package main

// RootCause возвращает самую глубокую ошибку цепочки:
// снимает слои через errors.Unwrap, пока они не закончатся.
// Для nil возвращает nil.
func RootCause(err error) error {
	// TODO: guard для nil + цикл с errors.Unwrap (import "errors")
	return nil
}
