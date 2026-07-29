package main

import "testing"

func TestOldestSingle(t *testing.T) {
	got := Oldest([]Person{{Name: "Аня", Age: 20}})
	if got != "Аня" {
		t.Errorf("Oldest со слайсом из одного человека: получено %q, ожидалось \"Аня\" — единственный человек и есть самый старший", got)
	}
}

func TestOldestAtEnd(t *testing.T) {
	people := []Person{
		{Name: "Боря", Age: 25},
		{Name: "Вера", Age: 31},
		{Name: "Тимофей", Age: 74},
	}
	if got := Oldest(people); got != "Тимофей" {
		t.Errorf("Oldest: самый старший стоит в конце слайса; получено %q, ожидалось \"Тимофей\"", got)
	}
}

func TestOldestTieFirst(t *testing.T) {
	people := []Person{
		{Name: "Игорь", Age: 40},
		{Name: "Кира", Age: 40},
		{Name: "Лев", Age: 12},
	}
	if got := Oldest(people); got != "Игорь" {
		t.Errorf("Oldest при равных возрастах должен вернуть ПЕРВОГО из равных; получено %q, ожидалось \"Игорь\" — обновляйте лидера только при СТРОГО большем возрасте", got)
	}
}

func TestOldestAtStart(t *testing.T) {
	people := []Person{
		{Name: "Мира", Age: 90},
		{Name: "Нина", Age: 8},
		{Name: "Олег", Age: 33},
	}
	if got := Oldest(people); got != "Мира" {
		t.Errorf("Oldest: самый старший стоит первым в слайсе; получено %q, ожидалось \"Мира\"", got)
	}
}
