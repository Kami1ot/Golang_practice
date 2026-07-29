package main

// SafeDivide делит a на b, ловя панику деления на ноль через recover
// и превращая её в ошибку "паника: <значение>".
// Понадобится fmt.Errorf — добавьте import "fmt".
func SafeDivide(a, b int) (result int, err error) {
	// TODO: defer с recover, затем наивное деление
	return 0, nil
}
