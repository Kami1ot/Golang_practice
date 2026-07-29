package main

// CountLines считает строки в файле path через bufio.Scanner.
// Файл не открылся — верните 0 и ошибку (не потеряйте исходную!).
// Понадобятся os и bufio — добавьте импорты сами.
func CountLines(path string) (int, error) {
	// TODO: os.Open → defer Close → Scanner → счётчик → sc.Err()
	return 0, nil
}

// WriteLines записывает строки в файл path, каждую завершая \n.
// Существующий файл перезаписывается; пустой срез — пустой файл.
func WriteLines(path string, lines []string) error {
	// TODO: os.Create → defer Close → fmt.Fprintln для каждой строки
	return nil
}
