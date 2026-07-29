package main

import (
	"reflect"
	"testing"
)

func TestRankBasic(t *testing.T) {
	players := []Player{{"Юра", 70}, {"Аня", 90}, {"Боря", 90}}
	Rank(players)
	want := []Player{{"Аня", 90}, {"Боря", 90}, {"Юра", 70}}
	if !reflect.DeepEqual(players, want) {
		t.Errorf("Rank дал %v, ожидалось %v — очки по убыванию, при равенстве имена по алфавиту", players, want)
	}
}

func TestRankAllDifferent(t *testing.T) {
	players := []Player{{"Ваня", 10}, {"Оля", 100}, {"Марк", 55}}
	Rank(players)
	want := []Player{{"Оля", 100}, {"Марк", 55}, {"Ваня", 10}}
	if !reflect.DeepEqual(players, want) {
		t.Errorf("Rank дал %v, ожидалось %v", players, want)
	}
}

func TestRankAllEqualScores(t *testing.T) {
	players := []Player{{"Соня", 50}, {"Аня", 50}, {"Боря", 50}}
	Rank(players)
	want := []Player{{"Аня", 50}, {"Боря", 50}, {"Соня", 50}}
	if !reflect.DeepEqual(players, want) {
		t.Errorf("Rank при равных очках дал %v, ожидалось %v — работает только запасной ключ", players, want)
	}
}

func TestRankEdge(t *testing.T) {
	empty := []Player{}
	Rank(empty)
	if len(empty) != 0 {
		t.Errorf("Rank пустого слайса что-то в него добавил: %v", empty)
	}
	single := []Player{{"Один", 1}}
	Rank(single)
	if !reflect.DeepEqual(single, []Player{{"Один", 1}}) {
		t.Errorf("Rank([{Один 1}]) дал %v, ожидалось без изменений", single)
	}
}
