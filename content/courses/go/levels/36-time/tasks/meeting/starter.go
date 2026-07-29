package main

import "time"

// Layout — формат даты и времени начала встречи, например "2026-03-10 14:30".
const Layout = "2006-01-02 15:04"

// ParseMeeting разбирает начало встречи (строка формата Layout) и её
// длительность (строка для ParseDuration), возвращая моменты начала и конца.
// Любая ошибка разбора пробрасывается наверх, времена при этом — нулевые.
func ParseMeeting(date, dur string) (time.Time, time.Time, error) {
	// TODO: time.Parse(Layout, date) → начало; time.ParseDuration(dur) → длительность;
	// конец = начало.Add(длительность); ошибки не глотать!
	return time.Time{}, time.Time{}, nil
}

// IsWeekend сообщает, выпадает ли момент t на субботу или воскресенье.
func IsWeekend(t time.Time) bool {
	// TODO: t.Weekday() и сравнение с time.Saturday / time.Sunday
	return false
}
