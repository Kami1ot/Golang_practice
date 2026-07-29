package main

import (
	"bytes"
	"testing"
)

func TestReceiptTwoItems(t *testing.T) {
	var buf bytes.Buffer
	WriteReceipt(&buf, []Item{{Name: "хлеб", Price: 40}, {Name: "молоко", Price: 90}})
	want := "- хлеб: 40 ₽\n- молоко: 90 ₽\nитого: 130 ₽\n"
	if got := buf.String(); got != want {
		t.Errorf("чек:\n%q\nожидалось:\n%q", got, want)
	}
}

func TestReceiptSingle(t *testing.T) {
	var buf bytes.Buffer
	WriteReceipt(&buf, []Item{{Name: "кофе", Price: 250}})
	want := "- кофе: 250 ₽\nитого: 250 ₽\n"
	if got := buf.String(); got != want {
		t.Errorf("чек из одного товара:\n%q\nожидалось:\n%q", got, want)
	}
}

func TestReceiptEmpty(t *testing.T) {
	var buf bytes.Buffer
	WriteReceipt(&buf, []Item{})
	want := "итого: 0 ₽\n"
	if got := buf.String(); got != want {
		t.Errorf("пустой чек: %q, ожидалось %q — только итоговая строка", got, want)
	}
	buf.Reset()
	WriteReceipt(&buf, nil)
	if got := buf.String(); got != want {
		t.Errorf("nil-чек: %q, ожидалось %q", got, want)
	}
}

func TestReceiptZeroPrice(t *testing.T) {
	var buf bytes.Buffer
	WriteReceipt(&buf, []Item{{Name: "подарок", Price: 0}, {Name: "чай", Price: 60}})
	want := "- подарок: 0 ₽\n- чай: 60 ₽\nитого: 60 ₽\n"
	if got := buf.String(); got != want {
		t.Errorf("чек с бесплатным товаром:\n%q\nожидалось:\n%q", got, want)
	}
}
