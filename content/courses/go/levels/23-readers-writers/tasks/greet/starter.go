package main

import (
	"fmt"
	"io"
)

// WriteGreeting пишет в w строку "Привет, <name>!" без перевода строки.
func WriteGreeting(w io.Writer, name string) {
	// TODO: подставьте имя — Fprintf понимает те же глаголы, что Printf
	fmt.Fprint(w, "")
}
