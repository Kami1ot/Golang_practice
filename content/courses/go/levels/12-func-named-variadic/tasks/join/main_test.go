package main

import "testing"

func TestJoinEmpty(t *testing.T) {
	got := Join(", ")
	if got != "" {
		t.Errorf("Join(\", \") = %q, ожидалась пустая строка — вызов без строк-частей", got)
	}
}

func TestJoinSingle(t *testing.T) {
	got := Join("-", "go")
	if got != "go" {
		t.Errorf("Join(\"-\", \"go\") = %q, ожидалось \"go\" — единственная часть возвращается без разделителей", got)
	}
}

func TestJoinSeveral(t *testing.T) {
	got := Join(", ", "a", "b", "c")
	if got != "a, b, c" {
		t.Errorf("Join(\", \", \"a\", \"b\", \"c\") = %q, ожидалось \"a, b, c\" — разделитель ставится МЕЖДУ частями, без хвоста в конце", got)
	}
}

func TestJoinEmptySep(t *testing.T) {
	got := Join("", "го", "фер")
	if got != "гофер" {
		t.Errorf("Join(\"\", \"го\", \"фер\") = %q, ожидалось \"гофер\" — пустой разделитель просто склеивает части вплотную", got)
	}
}

func TestJoinCyrillic(t *testing.T) {
	got := Join(" и ", "красный", "зелёный", "синий")
	if got != "красный и зелёный и синий" {
		t.Errorf("Join(\" и \", \"красный\", \"зелёный\", \"синий\") = %q, ожидалось \"красный и зелёный и синий\"", got)
	}
}
