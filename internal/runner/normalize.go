package runner

import "strings"

// Normalize приводит вывод к канонической форме для сравнения:
// CRLF -> LF, обрезка хвостовых пробелов/табов в каждой строке,
// обрезка завершающих переводов строки.
// Используется и для io-задач, и для quiz-вопросов типа "output".
func Normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}
