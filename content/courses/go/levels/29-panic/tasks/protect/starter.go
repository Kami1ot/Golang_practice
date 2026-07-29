package main

// Protect выполняет f, превращая любую её панику в ошибку "паника: <значение>".
// Спокойная работа f — nil. Понадобится fmt.Errorf — добавьте import "fmt".
func Protect(f func()) (err error) {
	// TODO: defer с recover, затем вызов f()
	return nil
}
