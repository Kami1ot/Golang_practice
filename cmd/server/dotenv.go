package main

import (
	"fmt"
	"os"
	"strings"
)

// loadDotEnv загружает простой файл KEY=VALUE, не перезаписывая уже заданное
// окружение. Это позволяет хранить локальный ключ в .env, а в production
// передавать переменные окружения обычным способом.
func loadDotEnv(path string) (bool, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for lineNo, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return false, fmt.Errorf("строка %d: ожидается KEY=VALUE", lineNo+1)
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return false, fmt.Errorf("строка %d: %w", lineNo+1, err)
			}
		}
	}
	return true, nil
}
