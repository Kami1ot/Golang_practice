package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CLIProvider гоняет наставника через Claude Code в headless-режиме
// (`claude -p`) — работает на подписке пользователя, без ключа API.
// Медленнее API (5–20 сек на ответ) и делит лимиты с работой в Claude Code.
type CLIProvider struct {
	bin   string
	model string
	sem   chan struct{} // CLI-сессии дорогие: один запрос за раз
}

// FindCLI ищет claude в PATH и типичных местах установки Windows.
func FindCLI() string {
	if p, err := exec.LookPath("claude"); err == nil {
		return p
	}
	home, _ := os.UserHomeDir()
	for _, candidate := range []string{
		filepath.Join(home, ".local", "bin", "claude.exe"),
		filepath.Join(home, ".local", "bin", "claude"),
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func NewCLIProvider(bin, model string) *CLIProvider {
	return &CLIProvider{bin: bin, model: model, sem: make(chan struct{}, 1)}
}

func (p *CLIProvider) Name() string { return "cli" }

func (p *CLIProvider) Reply(ctx context.Context, system string, messages []Message) (string, error) {
	select {
	case p.sem <- struct{}{}:
		defer func() { <-p.sem }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Headless-режим не хранит наш диалог — передаём транскрипт одним промптом.
	var sb strings.Builder
	sb.WriteString("Диалог с учеником:\n\n")
	for _, m := range messages {
		if m.Role == "assistant" {
			sb.WriteString("Наставник: ")
		} else {
			sb.WriteString("Ученик: ")
		}
		sb.WriteString(m.Content)
		sb.WriteString("\n\n")
	}
	sb.WriteString("Ответь на последнее сообщение ученика (только текст ответа, без префикса «Наставник:»).")

	args := []string{
		"-p", sb.String(),
		"--append-system-prompt", system,
		"--output-format", "json",
		"--max-turns", "1",
	}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	cmd := exec.CommandContext(ctx, p.bin, args...)
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return "", fmt.Errorf("наставник думал слишком долго, попробуйте ещё раз")
	}
	if err != nil {
		detail := ""
		if ee, ok := err.(*exec.ExitError); ok {
			detail = ": " + strings.TrimSpace(string(ee.Stderr))
		}
		return "", fmt.Errorf("claude CLI%s", detail)
	}

	// --output-format json → {"type":"result","result":"текст",...}
	var parsed struct {
		Result  string `json:"result"`
		IsError bool   `json:"is_error"`
	}
	if jsonErr := json.Unmarshal(out, &parsed); jsonErr == nil && parsed.Result != "" {
		if parsed.IsError {
			return "", fmt.Errorf("claude CLI: %s", parsed.Result)
		}
		return parsed.Result, nil
	}
	// Формат вывода мог поменяться между версиями — деградируем в сырой текст.
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return "", fmt.Errorf("пустой ответ claude CLI")
	}
	return raw, nil
}
