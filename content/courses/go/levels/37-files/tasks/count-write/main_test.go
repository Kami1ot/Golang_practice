package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("не удалось подготовить временный файл: %v", err)
	}
	return path
}

func TestCountLinesBasic(t *testing.T) {
	path := writeTemp(t, "альфа\nбета\nгамма\n")
	got, err := CountLines(path)
	if err != nil {
		t.Fatalf("CountLines существующего файла вернула ошибку %v, ожидался nil", err)
	}
	if got != 3 {
		t.Errorf("CountLines файла из трёх строк = %d, ожидалось 3", got)
	}
}

func TestCountLinesEmpty(t *testing.T) {
	path := writeTemp(t, "")
	got, err := CountLines(path)
	if err != nil {
		t.Fatalf("CountLines пустого файла вернула ошибку %v, ожидался nil", err)
	}
	if got != 0 {
		t.Errorf("CountLines пустого файла = %d, ожидалось 0", got)
	}
}

func TestCountLinesNoTrailingNewline(t *testing.T) {
	path := writeTemp(t, "раз\nдва")
	got, err := CountLines(path)
	if err != nil {
		t.Fatalf("CountLines вернула ошибку %v, ожидался nil", err)
	}
	if got != 2 {
		t.Errorf("CountLines файла «раз\\nдва» = %d, ожидалось 2 — последняя строка без \\n тоже считается", got)
	}
}

func TestCountLinesMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "такого-файла-нет.txt")
	got, err := CountLines(path)
	if err == nil {
		t.Fatalf("CountLines несуществующего файла вернула nil-ошибку, ожидалась ошибка открытия")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ошибка %v не проходит errors.Is(err, fs.ErrNotExist) — возвращайте исходную ошибку os.Open (как есть или обёрнутой через %%w), а не свою", err)
	}
	if got != 0 {
		t.Errorf("CountLines несуществующего файла = %d, ожидалось 0", got)
	}
}

func TestWriteLinesBasic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	if err := WriteLines(path, []string{"один", "два", "три"}); err != nil {
		t.Fatalf("WriteLines вернула ошибку %v, ожидался nil", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("файл после WriteLines не читается: %v — файл вообще создан?", err)
	}
	want := "один\nдва\nтри\n"
	if string(data) != want {
		t.Errorf("содержимое файла = %q, ожидалось %q — каждая строка должна завершаться \\n", string(data), want)
	}
}

func TestWriteLinesEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.txt")
	if err := WriteLines(path, nil); err != nil {
		t.Fatalf("WriteLines с пустым срезом вернула ошибку %v, ожидался nil", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("файл после WriteLines(nil) не читается: %v — пустой срез всё равно должен создать файл", err)
	}
	if len(data) != 0 {
		t.Errorf("содержимое файла = %q, ожидался пустой файл", string(data))
	}
}

func TestWriteLinesOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rewrite.txt")
	if err := WriteLines(path, []string{"старое", "содержимое", "на три строки"}); err != nil {
		t.Fatalf("первая запись: %v", err)
	}
	if err := WriteLines(path, []string{"новое"}); err != nil {
		t.Fatalf("вторая запись: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("файл не читается: %v", err)
	}
	if string(data) != "новое\n" {
		t.Errorf("после повторной записи содержимое = %q, ожидалось %q — файл должен перезаписываться, а не дополняться", string(data), "новое\n")
	}
}

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roundtrip.txt")
	lines := []string{"go", "любит", "простые решения", "  и пробелы тоже"}
	if err := WriteLines(path, lines); err != nil {
		t.Fatalf("WriteLines: %v", err)
	}
	got, err := CountLines(path)
	if err != nil {
		t.Fatalf("CountLines после WriteLines: %v", err)
	}
	if got != len(lines) {
		t.Errorf("записали %d строк, CountLines насчитала %d — функции должны сходиться", len(lines), got)
	}
}
