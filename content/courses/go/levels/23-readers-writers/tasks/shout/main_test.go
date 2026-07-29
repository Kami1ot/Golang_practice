package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestShoutCyrillic(t *testing.T) {
	var buf bytes.Buffer
	err := Shout(strings.NewReader("говорите громче"), &buf)
	if err != nil {
		t.Fatalf("Shout вернула ошибку %v, ожидался nil", err)
	}
	if got := buf.String(); got != "ГОВОРИТЕ ГРОМЧЕ" {
		t.Errorf("в буфере %q, ожидалось %q", got, "ГОВОРИТЕ ГРОМЧЕ")
	}
}

func TestShoutMixed(t *testing.T) {
	var buf bytes.Buffer
	if err := Shout(strings.NewReader("Go 1.26 — сила!"), &buf); err != nil {
		t.Fatalf("Shout вернула ошибку %v, ожидался nil", err)
	}
	if got := buf.String(); got != "GO 1.26 — СИЛА!" {
		t.Errorf("в буфере %q, ожидалось %q — цифры и знаки не меняются", got, "GO 1.26 — СИЛА!")
	}
}

func TestShoutEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := Shout(strings.NewReader(""), &buf); err != nil {
		t.Fatalf("Shout на пустом источнике вернула ошибку %v, ожидался nil", err)
	}
	if got := buf.String(); got != "" {
		t.Errorf("в буфере %q, ожидалась пустая строка", got)
	}
}

// brokenReader всегда падает с одной и той же ошибкой.
type brokenReader struct{}

var errBroken = errors.New("ридер сломан")

func (brokenReader) Read(p []byte) (int, error) {
	return 0, errBroken
}

func TestShoutReadError(t *testing.T) {
	var buf bytes.Buffer
	err := Shout(brokenReader{}, &buf)
	if err == nil {
		t.Fatalf("Shout со сломанным ридером вернула nil, ожидалась ошибка")
	}
	if !errors.Is(err, errBroken) {
		t.Errorf("Shout вернула %v, ожидалась исходная ошибка ридера — не подменяйте её своей", err)
	}
	if buf.Len() != 0 {
		t.Errorf("при ошибке чтения в буфер записано %q, ожидалась пустота", buf.String())
	}
}
