// Package chat — провайдеры LLM для чата-наставника.
// Реализации: api (Anthropic API через официальный SDK) и cli (Claude Code headless).
package chat

import "context"

type Message struct {
	Role    string // "user" | "assistant"
	Content string
}

// Provider отвечает на диалог одним сообщением наставника.
type Provider interface {
	Reply(ctx context.Context, system string, messages []Message) (string, error)
	// Name — человекочитаемое имя провайдера для логов ("api", "cli").
	Name() string
}
