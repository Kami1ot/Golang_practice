package main

import (
	"testing"
	"time"
)

func TestParseMeetingOK(t *testing.T) {
	start, end, err := ParseMeeting("2026-03-10 14:30", "1h30m")
	if err != nil {
		t.Fatalf("ParseMeeting корректного ввода вернула ошибку: %v", err)
	}
	wantStart := time.Date(2026, time.March, 10, 14, 30, 0, 0, time.UTC)
	wantEnd := time.Date(2026, time.March, 10, 16, 0, 0, 0, time.UTC)
	if !start.Equal(wantStart) {
		t.Errorf("начало встречи = %v, ожидалось %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("конец встречи = %v, ожидалось %v (начало + 1h30m)", end, wantEnd)
	}
}

func TestParseMeetingShort(t *testing.T) {
	start, end, err := ParseMeeting("2026-07-17 09:00", "45m")
	if err != nil {
		t.Fatalf("ParseMeeting(\"2026-07-17 09:00\", \"45m\") вернула ошибку: %v", err)
	}
	if got := end.Sub(start); got != 45*time.Minute {
		t.Errorf("длительность встречи = %v, ожидалось 45m0s", got)
	}
}

func TestParseMeetingCrossMidnight(t *testing.T) {
	_, end, err := ParseMeeting("2026-12-31 23:30", "45m")
	if err != nil {
		t.Fatalf("ParseMeeting новогодней встречи вернула ошибку: %v", err)
	}
	want := time.Date(2027, time.January, 1, 0, 15, 0, 0, time.UTC)
	if !end.Equal(want) {
		t.Errorf("конец встречи через полночь = %v, ожидалось %v — Add обязан перенести и день, и год", end, want)
	}
}

func TestParseMeetingBadDate(t *testing.T) {
	start, end, err := ParseMeeting("10.03.2026 14:30", "1h")
	if err == nil {
		t.Fatal("дата «10.03.2026 14:30» не подходит под Layout — ожидалась ошибка, получен nil")
	}
	if !start.IsZero() || !end.IsZero() {
		t.Errorf("при ошибке разбора времена должны быть нулевыми, получено start=%v end=%v", start, end)
	}
}

func TestParseMeetingBadDuration(t *testing.T) {
	start, end, err := ParseMeeting("2026-03-10 14:30", "полтора часа")
	if err == nil {
		t.Fatal("длительность «полтора часа» не разбирается ParseDuration — ожидалась ошибка, получен nil")
	}
	if !start.IsZero() || !end.IsZero() {
		t.Errorf("при ошибке разбора времена должны быть нулевыми, получено start=%v end=%v", start, end)
	}
}

func TestIsWeekend(t *testing.T) {
	cases := []struct {
		name string
		t    time.Time
		want bool
	}{
		{"суббота", time.Date(2026, time.March, 14, 10, 0, 0, 0, time.UTC), true},
		{"воскресенье", time.Date(2026, time.March, 15, 23, 59, 0, 0, time.UTC), true},
		{"понедельник", time.Date(2026, time.March, 16, 0, 0, 0, 0, time.UTC), false},
		{"пятница", time.Date(2026, time.March, 13, 18, 0, 0, 0, time.UTC), false},
	}
	for _, c := range cases {
		if got := IsWeekend(c.t); got != c.want {
			t.Errorf("IsWeekend(%s, %s) = %v, ожидалось %v",
				c.t.Format("2006-01-02"), c.name, got, c.want)
		}
	}
}
