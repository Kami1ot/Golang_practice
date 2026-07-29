package main

import (
	"testing"
	"time"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestAgeBirthdayPassed(t *testing.T) {
	if got := Age(date(2000, time.May, 15), date(2026, time.July, 17)); got != 26 {
		t.Errorf("Age(2000-05-15, 2026-07-17) = %d, ожидалось 26 — майский день рождения давно был", got)
	}
}

func TestAgeBirthdayAhead(t *testing.T) {
	if got := Age(date(2000, time.September, 1), date(2026, time.July, 17)); got != 25 {
		t.Errorf("Age(2000-09-01, 2026-07-17) = %d, ожидалось 25 — сентябрь ещё не наступил", got)
	}
}

func TestAgeBirthdayToday(t *testing.T) {
	if got := Age(date(2000, time.July, 17), date(2026, time.July, 17)); got != 26 {
		t.Errorf("Age(2000-07-17, 2026-07-17) = %d, ожидалось 26 — день рождения сегодня уже засчитан", got)
	}
}

func TestAgeBirthdayTomorrow(t *testing.T) {
	if got := Age(date(2000, time.July, 18), date(2026, time.July, 17)); got != 25 {
		t.Errorf("Age(2000-07-18, 2026-07-17) = %d, ожидалось 25 — до дня рождения ещё день", got)
	}
}

func TestAgeYearBoundary(t *testing.T) {
	if got := Age(date(1990, time.December, 31), date(2026, time.January, 1)); got != 35 {
		t.Errorf("Age(1990-12-31, 2026-01-01) = %d, ожидалось 35 — новый год не значит новый полный год", got)
	}
}

func TestAgeLessThanYear(t *testing.T) {
	if got := Age(date(2025, time.June, 1), date(2026, time.May, 31)); got != 0 {
		t.Errorf("Age(2025-06-01, 2026-05-31) = %d, ожидалось 0 — год ещё не полный", got)
	}
}

func TestAgeLeapBirthday(t *testing.T) {
	// Рождённый 29 февраля: в невисокосный год его «день рождения»
	// нормализуется AddDate на 1 марта.
	if got := Age(date(2004, time.February, 29), date(2026, time.February, 28)); got != 21 {
		t.Errorf("Age(2004-02-29, 2026-02-28) = %d, ожидалось 21 — 28 февраля год ещё не полный", got)
	}
	if got := Age(date(2004, time.February, 29), date(2026, time.March, 1)); got != 22 {
		t.Errorf("Age(2004-02-29, 2026-03-01) = %d, ожидалось 22 — 1 марта год засчитан", got)
	}
}
