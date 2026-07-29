package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func prepareSrc(t *testing.T, content string) (srcPath, dstPath string) {
	t.Helper()
	dir := t.TempDir()
	srcPath = filepath.Join(dir, "src.txt")
	dstPath = filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(srcPath, []byte(content), 0644); err != nil {
		t.Fatalf("не удалось подготовить исходный файл: %v", err)
	}
	return srcPath, dstPath
}

func TestGrepBasic(t *testing.T) {
	src, dst := prepareSrc(t, "[info] старт\n[error] диск полон\n[info] запрос\n[error] снова диск\n")
	got, err := Grep(src, dst, "[error]")
	if err != nil {
		t.Fatalf("Grep вернула ошибку %v, ожидался nil", err)
	}
	if got != 2 {
		t.Errorf("Grep нашла %d строк, ожидалось 2", got)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("файл-приёмник не читается: %v — он вообще создан?", err)
	}
	want := "[error] диск полон\n[error] снова диск\n"
	if string(data) != want {
		t.Errorf("содержимое приёмника = %q, ожидалось %q — только совпавшие строки, каждая с \\n; Flush не забыли?", string(data), want)
	}
}

func TestGrepNoMatches(t *testing.T) {
	src, dst := prepareSrc(t, "тишина\nи покой\n")
	got, err := Grep(src, dst, "гроза")
	if err != nil {
		t.Fatalf("Grep без совпадений вернула ошибку %v, ожидался nil", err)
	}
	if got != 0 {
		t.Errorf("Grep без совпадений = %d, ожидалось 0", got)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("приёмник должен быть создан даже без совпадений, а он не читается: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("приёмник без совпадений должен быть пуст, а в нём %q", string(data))
	}
}

func TestGrepCyrillic(t *testing.T) {
	src, dst := prepareSrc(t, "кот на крыше\nсобака во дворе\nкоты повсюду\nпёс молчит")
	got, err := Grep(src, dst, "кот")
	if err != nil {
		t.Fatalf("Grep вернула ошибку %v, ожидался nil", err)
	}
	if got != 2 {
		t.Errorf("Grep по подстроке «кот» = %d, ожидалось 2 — совпадение ищется в любом месте строки", got)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("файл-приёмник не читается: %v", err)
	}
	want := "кот на крыше\nкоты повсюду\n"
	if string(data) != want {
		t.Errorf("содержимое приёмника = %q, ожидалось %q — последняя строка источника без \\n тоже проверяется", string(data), want)
	}
}

func TestGrepMissingSource(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "нет-такого.log")
	dst := filepath.Join(dir, "dst.txt")
	got, err := Grep(src, dst, "что угодно")
	if err == nil {
		t.Fatalf("Grep несуществующего источника вернула nil-ошибку, ожидалась ошибка открытия")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ошибка %v не проходит errors.Is(err, fs.ErrNotExist) — возвращайте исходную ошибку os.Open (как есть или через %%w), а не свою", err)
	}
	if got != 0 {
		t.Errorf("Grep несуществующего источника = %d, ожидалось 0", got)
	}
}

func TestGrepOverwrite(t *testing.T) {
	src, dst := prepareSrc(t, "a1\nb2\na3\n")
	if err := os.WriteFile(dst, []byte("старое содержимое приёмника\nна две строки\n"), 0644); err != nil {
		t.Fatalf("не удалось подготовить приёмник: %v", err)
	}
	got, err := Grep(src, dst, "a")
	if err != nil {
		t.Fatalf("Grep вернула ошибку %v, ожидался nil", err)
	}
	if got != 2 {
		t.Errorf("Grep = %d, ожидалось 2", got)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("файл-приёмник не читается: %v", err)
	}
	want := "a1\na3\n"
	if string(data) != want {
		t.Errorf("содержимое приёмника = %q, ожидалось %q — существующий приёмник перезаписывается, а не дополняется", string(data), want)
	}
}
