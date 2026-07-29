package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestSumBasic(t *testing.T) {
	data := `{"kind":"click","n":2}
{"kind":"view","n":1}
{"kind":"click","n":3}`
	got, err := SumEvents(strings.NewReader(data))
	if err != nil {
		t.Fatalf("SumEvents(3 события) вернула ошибку: %v", err)
	}
	want := map[string]int{"click": 5, "view": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SumEvents(click:2, view:1, click:3) = %v, ожидалось %v", got, want)
	}
}

func TestSumNegative(t *testing.T) {
	data := `{"kind":"score","n":10}
{"kind":"score","n":-4}`
	got, err := SumEvents(strings.NewReader(data))
	if err != nil {
		t.Fatalf("SumEvents вернула ошибку: %v", err)
	}
	if got["score"] != 6 {
		t.Errorf("сумма score = %d, ожидалось 6 (10 + (-4)) — отрицательные n складываются как обычно", got["score"])
	}
}

func TestSumEmpty(t *testing.T) {
	got, err := SumEvents(strings.NewReader(""))
	if err != nil {
		t.Fatalf("SumEvents(пустой поток) вернула ошибку: %v — пустой поток не ошибка", err)
	}
	if len(got) != 0 {
		t.Errorf("SumEvents(пустой поток) = %v, ожидалась карта без событий", got)
	}
}

func TestSumPretty(t *testing.T) {
	data := `{
  "kind": "загрузка",
  "n": 7
}
{ "kind": "загрузка",  "n": 1 }`
	got, err := SumEvents(strings.NewReader(data))
	if err != nil {
		t.Fatalf("SumEvents(отформатированные события) вернула ошибку: %v — декодеру отступы не мешают", err)
	}
	if got["загрузка"] != 8 {
		t.Errorf("сумма «загрузка» = %d, ожидалось 8", got["загрузка"])
	}
}

func TestSumBadJSON(t *testing.T) {
	data := `{"kind":"ok","n":1}
{кривой json`
	_, err := SumEvents(strings.NewReader(data))
	if err == nil {
		t.Errorf("SumEvents(битый поток) вернула nil-ошибку — ошибку Decode нужно вернуть наружу")
	}
}
