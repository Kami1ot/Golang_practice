package main

import "testing"

func TestMoveOnce(t *testing.T) {
	p := Point{X: 1, Y: 2}
	Move(&p, 3, 4)
	if p.X != 4 || p.Y != 6 {
		t.Errorf("после Move(&p, 3, 4) из точки (1, 2) получилось (%d, %d), ожидалось (4, 6). Move должна менять поля через указатель", p.X, p.Y)
	}
}

func TestMoveAccumulates(t *testing.T) {
	p := Point{}
	Move(&p, 1, 1)
	Move(&p, 2, 3)
	if p.X != 3 || p.Y != 4 {
		t.Errorf("после Move(&p, 1, 1) и Move(&p, 2, 3) из точки (0, 0) получилось (%d, %d), ожидалось (3, 4) — сдвиги должны накапливаться, а не затирать друг друга", p.X, p.Y)
	}
}

func TestMoveNegative(t *testing.T) {
	p := Point{X: 3, Y: 3}
	Move(&p, -5, -1)
	if p.X != -2 || p.Y != 2 {
		t.Errorf("после Move(&p, -5, -1) из точки (3, 3) получилось (%d, %d), ожидалось (-2, 2) — отрицательные смещения тоже работают", p.X, p.Y)
	}
}

func TestResetAfterMoves(t *testing.T) {
	p := Point{X: 5, Y: -2}
	Move(&p, 10, 10)
	Reset(&p)
	if p.X != 0 || p.Y != 0 {
		t.Errorf("после Reset(&p) точка осталась (%d, %d), ожидалось (0, 0) — Reset должна обнулять ОБЕ координаты оригинала", p.X, p.Y)
	}
}
