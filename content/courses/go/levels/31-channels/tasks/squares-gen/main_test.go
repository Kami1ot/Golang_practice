package main

import (
	"reflect"
	"testing"
	"time"
)

func collect(t *testing.T, ch <-chan int, expectMax int) []int {
	t.Helper()
	var got []int
	timeout := time.After(3 * time.Second)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return got
			}
			got = append(got, v)
			if len(got) > expectMax {
				t.Fatalf("канал отдал больше %d значений и не закрылся — close после последней отправки!", expectMax)
			}
		case <-timeout:
			t.Fatalf("канал не закрылся: собрано %v, а range ждёт следующее значение вечно", got)
		}
	}
}

func TestSquaresBasic(t *testing.T) {
	got := collect(t, Squares(4), 4)
	want := []int{1, 4, 9, 16}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Squares(4) отдал %v, ожидалось %v — квадраты по возрастанию", got, want)
	}
}

func TestSquaresOne(t *testing.T) {
	got := collect(t, Squares(1), 1)
	if !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("Squares(1) отдал %v, ожидалось [1]", got)
	}
}

func TestSquaresZero(t *testing.T) {
	got := collect(t, Squares(0), 0)
	if len(got) != 0 {
		t.Errorf("Squares(0) отдал %v, ожидался пустой закрытый канал", got)
	}
}

func TestSquaresClosedAfterDrain(t *testing.T) {
	ch := Squares(2)
	collect(t, ch, 2)
	v, ok := <-ch
	if ok {
		t.Errorf("после всех значений канал должен быть закрыт, а из него пришло ещё %d", v)
	}
	if v != 0 {
		t.Errorf("приём из закрытого канала = %d, ожидалось нулевое значение 0", v)
	}
}
