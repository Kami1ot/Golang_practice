package main

import "testing"

func TestFinalScoreBasic(t *testing.T) {
	if got := FinalScore(5); got != 20 {
		t.Errorf("FinalScore(5) = %d, ожидалось 20 (5*2 = 10, плюс бонус 10 из defer)", got)
	}
}

func TestFinalScoreZero(t *testing.T) {
	if got := FinalScore(0); got != 10 {
		t.Errorf("FinalScore(0) = %d, ожидалось 10 — база нулевая, но бонус 10 добавляется всегда", got)
	}
}

func TestFinalScoreNegative(t *testing.T) {
	if got := FinalScore(-3); got != 4 {
		t.Errorf("FinalScore(-3) = %d, ожидалось 4 (-3*2 = -6, плюс бонус 10)", got)
	}
}

func TestFinalScoreOne(t *testing.T) {
	if got := FinalScore(1); got != 12 {
		t.Errorf("FinalScore(1) = %d, ожидалось 12 (1*2 = 2, плюс бонус 10)", got)
	}
}

func TestFinalScoreLarge(t *testing.T) {
	if got := FinalScore(100); got != 210 {
		t.Errorf("FinalScore(100) = %d, ожидалось 210 (100*2 = 200, плюс бонус 10)", got)
	}
}
