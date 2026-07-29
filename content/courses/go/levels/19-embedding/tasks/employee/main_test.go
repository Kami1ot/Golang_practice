package main

import "testing"

func TestPersonGreet(t *testing.T) {
	p := Person{Name: "Ира", Age: 30}
	want := "Привет, я Ира"
	if got := p.Greet(); got != want {
		t.Errorf("Person.Greet() = %q, ожидалось %q", got, want)
	}
}

func TestEmployeeGreetShadows(t *testing.T) {
	e := Employee{Person: Person{Name: "Ира", Age: 30}, Company: "Гофер-софт"}
	want := "Привет, я Ира из Гофер-софт"
	if got := e.Greet(); got != want {
		t.Errorf("e.Greet() = %q, ожидалось %q — собственный метод Employee должен перекрывать продвинутый Person.Greet и добавлять компанию", got, want)
	}
}

func TestInnerGreetViaFullName(t *testing.T) {
	e := Employee{Person: Person{Name: "Макс", Age: 45}, Company: "Слайс и партнёры"}
	want := "Привет, я Макс"
	if got := e.Person.Greet(); got != want {
		t.Errorf("e.Person.Greet() = %q, ожидалось %q — по полному имени должен вызываться внутренний метод Person, без упоминания компании", got, want)
	}
}

func TestEmployeeCard(t *testing.T) {
	e := Employee{Person: Person{Name: "Ира", Age: 30}, Company: "Гофер-софт"}
	want := "Ира, 30 лет, работает в Гофер-софт"
	if got := e.Card(); got != want {
		t.Errorf("e.Card() = %q, ожидалось %q — имя и возраст берутся из продвинутых полей Person", got, want)
	}
}

func TestEmployeeCardOtherData(t *testing.T) {
	e := Employee{Person: Person{Name: "Макс", Age: 45}, Company: "Слайс и партнёры"}
	want := "Макс, 45 лет, работает в Слайс и партнёры"
	if got := e.Card(); got != want {
		t.Errorf("e.Card() = %q, ожидалось %q", got, want)
	}
}
