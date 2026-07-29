package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("# comment\nDOTENV_TEST_VALUE=loaded\nDOTENV_TEST_QUOTED=\"two words\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOTENV_TEST_VALUE", "environment")
	loaded, err := loadDotEnv(path)
	if err != nil || !loaded {
		t.Fatalf("loadDotEnv = %v, %v", loaded, err)
	}
	if got := os.Getenv("DOTENV_TEST_VALUE"); got != "environment" {
		t.Fatalf("environment was overwritten: %q", got)
	}
	if got := os.Getenv("DOTENV_TEST_QUOTED"); got != "two words" {
		t.Fatalf("quoted value = %q", got)
	}
}
