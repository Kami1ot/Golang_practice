package main

import (
	"strings"
	"testing"
)

func TestParseAgeOK(t *testing.T) {
	got, err := ParseAge("33")
	if err != nil {
		t.Fatalf("ParseAge(\"33\") вернула ошибку %v, ожидался nil", err)
	}
	if got != 33 {
		t.Errorf("ParseAge(\"33\") = %d, ожидалось 33", got)
	}
}

func TestParseAgeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int
	}{{"0", 0}, {"150", 150}} {
		got, err := ParseAge(tc.in)
		if err != nil {
			t.Errorf("ParseAge(%q) вернула ошибку %v — границы диапазона входят в норму", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseAge(%q) = %d, ожидалось %d", tc.in, got, tc.want)
		}
	}
}

func TestParseAgeNotNumber(t *testing.T) {
	got, err := ParseAge("сто")
	if err == nil {
		t.Fatalf("ParseAge(\"сто\") вернула nil, ожидалась проброшенная ошибка strconv.Atoi")
	}
	if !strings.Contains(err.Error(), "invalid syntax") {
		t.Errorf("ошибка %q не похожа на проброшенную от strconv.Atoi — пробрасывайте её как есть, не подменяя", err.Error())
	}
	if got != 0 {
		t.Errorf("при ошибке результат = %d, ожидалось 0", got)
	}
}

func TestParseAgeOutOfRange(t *testing.T) {
	_, err := ParseAge("200")
	if err == nil {
		t.Fatalf("ParseAge(\"200\") вернула nil, ожидалась ошибка диапазона")
	}
	if err.Error() != "возраст 200 вне диапазона [0, 150]" {
		t.Errorf("текст ошибки %q, ожидалось %q", err.Error(), "возраст 200 вне диапазона [0, 150]")
	}
	if _, err := ParseAge("-5"); err == nil {
		t.Errorf("ParseAge(\"-5\") вернула nil, ожидалась ошибка — отрицательный возраст вне диапазона")
	}
}
