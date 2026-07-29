package main

import (
	"io"
)

// Shout читает всё из r, переводит в верхний регистр и пишет в w.
// Ошибку чтения или записи возвращает вызывающему; при успехе — nil.
func Shout(r io.Reader, w io.Writer) error {
	// TODO: io.ReadAll → strings.ToUpper → запись в w
	_ = r
	_ = w
	return nil
}
