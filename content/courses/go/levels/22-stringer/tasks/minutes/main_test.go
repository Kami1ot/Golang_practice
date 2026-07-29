package main

import (
	"fmt"
	"testing"
)

var _ fmt.Stringer = Minutes(0)

func TestLessThanHour(t *testing.T) {
	cases := map[Minutes]string{
		0:  "0м",
		1:  "1м",
		45: "45м",
		59: "59м",
	}
	for m, want := range cases {
		if got := m.String(); got != want {
			t.Errorf("Minutes(%d).String() = %q, ожидалось %q — до часа печатаем только минуты", int(m), got, want)
		}
	}
}

func TestHoursAndMinutes(t *testing.T) {
	cases := map[Minutes]string{
		60:  "1ч 00м",
		61:  "1ч 01м",
		125: "2ч 05м",
		599: "9ч 59м",
		600: "10ч 00м",
	}
	for m, want := range cases {
		if got := m.String(); got != want {
			t.Errorf("Minutes(%d).String() = %q, ожидалось %q — минуты дополняются нулём до двух цифр", int(m), got, want)
		}
	}
}

func TestFmtUsesStringer(t *testing.T) {
	if got := fmt.Sprintf("осталось %v", Minutes(90)); got != "осталось 1ч 30м" {
		t.Errorf("Sprintf(%%v, Minutes(90)) = %q, ожидалось %q", got, "осталось 1ч 30м")
	}
}
