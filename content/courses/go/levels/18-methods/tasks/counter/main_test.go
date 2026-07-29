package main

import "testing"

func TestCounterIncSeries(t *testing.T) {
	var c Counter
	c.Inc()
	c.Inc()
	c.Inc()
	if got := c.Value(); got != 3 {
		t.Errorf("после трёх Inc() счётчик = %d, ожидалось 3 — проверьте, что Inc действительно меняет c.value", got)
	}
}

func TestCounterAddPositive(t *testing.T) {
	var c Counter
	c.Add(5)
	if got := c.Value(); got != 5 {
		t.Errorf("после Add(5) счётчик = %d, ожидалось 5", got)
	}
}

func TestCounterAddNegative(t *testing.T) {
	var c Counter
	c.Add(10)
	c.Add(-4)
	if got := c.Value(); got != 6 {
		t.Errorf("после Add(10) и Add(-4) счётчик = %d, ожидалось 6 — отрицательный n должен уменьшать счётчик", got)
	}
}

func TestCounterValueDoesNotMutate(t *testing.T) {
	var c Counter
	c.Inc()
	_ = c.Value()
	_ = c.Value()
	if got := c.Value(); got != 1 {
		t.Errorf("после Inc() и нескольких чтений Value() счётчик = %d, ожидалось 1 — Value не должен менять состояние", got)
	}
}

func TestCounterZero(t *testing.T) {
	var c Counter
	if got := c.Value(); got != 0 {
		t.Errorf("Value() нового счётчика = %d, ожидалось 0", got)
	}
}

func TestCounterMixed(t *testing.T) {
	var c Counter
	c.Inc()
	c.Add(3)
	c.Inc()
	if got := c.Value(); got != 5 {
		t.Errorf("после Inc(), Add(3), Inc() счётчик = %d, ожидалось 5", got)
	}
}
