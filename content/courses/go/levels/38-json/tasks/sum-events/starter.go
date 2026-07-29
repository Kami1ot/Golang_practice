package main

import "io"

// Event — одно событие потока: {"kind":"...","n":...}
type Event struct {
	Kind string `json:"kind"`
	N    int    `json:"n"`
}

// SumEvents читает JSON-события из r до конца потока
// и возвращает карту «kind → сумма n этого вида».
func SumEvents(r io.Reader) (map[string]int, error) {
	// TODO: json.NewDecoder(r) + цикл for dec.More() { ... }
	return nil, nil
}
