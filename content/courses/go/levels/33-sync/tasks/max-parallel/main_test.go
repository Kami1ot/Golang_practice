package main

import "testing"

func TestMaxParallelBasic(t *testing.T) {
	rows := [][]int{{3, 1, 4}, {1, 5, 9}, {2, 6, 5}}
	if got := MaxParallel(rows); got != 9 {
		t.Errorf("MaxParallel = %d, ожидалось 9", got)
	}
}

func TestMaxParallelNegative(t *testing.T) {
	rows := [][]int{{-7, -3}, {-5, -10}}
	if got := MaxParallel(rows); got != -3 {
		t.Errorf("MaxParallel на отрицательной матрице = %d, ожидалось -3 — стартовать лидера с нуля нельзя", got)
	}
}

func TestMaxParallelSingleRow(t *testing.T) {
	if got := MaxParallel([][]int{{42}}); got != 42 {
		t.Errorf("MaxParallel([[42]]) = %d, ожидалось 42", got)
	}
}

func TestMaxParallelEmpty(t *testing.T) {
	if got := MaxParallel([][]int{}); got != 0 {
		t.Errorf("MaxParallel пустой матрицы = %d, ожидалось 0", got)
	}
	if got := MaxParallel(nil); got != 0 {
		t.Errorf("MaxParallel(nil) = %d, ожидалось 0", got)
	}
}

func TestMaxParallelManyRows(t *testing.T) {
	rows := make([][]int, 100)
	for i := range rows {
		rows[i] = []int{i, i - 50, i + 1}
	}
	// максимум: последняя строка, i=99 → 100
	if got := MaxParallel(rows); got != 100 {
		t.Errorf("MaxParallel на 100 строках = %d, ожидалось 100 — все ли горутины дождались и не потерялись ли обновления?", got)
	}
}
