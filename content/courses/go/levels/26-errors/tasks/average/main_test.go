package main

import (
	"math"
	"testing"
)

func TestAverageOK(t *testing.T) {
	got, err := Average([]int{4, 5})
	if err != nil {
		t.Fatalf("Average([4, 5]) вернула ошибку %v, ожидался nil", err)
	}
	if math.Abs(got-4.5) > 1e-9 {
		t.Errorf("Average([4, 5]) = %v, ожидалось 4.5 — не потеряйте дробную часть при делении", got)
	}
}

func TestAverageSingle(t *testing.T) {
	got, err := Average([]int{7})
	if err != nil {
		t.Fatalf("Average([7]) вернула ошибку %v, ожидался nil", err)
	}
	if math.Abs(got-7) > 1e-9 {
		t.Errorf("Average([7]) = %v, ожидалось 7", got)
	}
}

func TestAverageEmpty(t *testing.T) {
	got, err := Average([]int{})
	if err == nil {
		t.Fatalf("Average пустого слайса вернула nil, ожидалась ошибка")
	}
	if err.Error() != "нет данных" {
		t.Errorf("текст ошибки %q, ожидалось %q", err.Error(), "нет данных")
	}
	if got != 0 {
		t.Errorf("при ошибке результат = %v, ожидалось 0", got)
	}
	if _, err := Average(nil); err == nil {
		t.Errorf("Average(nil) вернула nil, ожидалась ошибка")
	}
}

func TestReportOK(t *testing.T) {
	if got := Report([]int{4, 5}); got != "среднее: 4.5" {
		t.Errorf("Report([4, 5]) = %q, ожидалось %q", got, "среднее: 4.5")
	}
	if got := Report([]int{10, 20, 30}); got != "среднее: 20.0" {
		t.Errorf("Report([10, 20, 30]) = %q, ожидалось %q — %%.1f печатает один знак всегда", got, "среднее: 20.0")
	}
}

func TestReportError(t *testing.T) {
	if got := Report(nil); got != "ошибка: нет данных" {
		t.Errorf("Report(nil) = %q, ожидалось %q", got, "ошибка: нет данных")
	}
}
