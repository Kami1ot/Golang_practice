// Package chat — провайдер OpenAI для чата-наставника.
package chat

import "context"

type Message struct {
	Role    string // "user" | "assistant"
	Content string
}

// Provider отвечает на диалог одним сообщением наставника.
type Provider interface {
	Reply(ctx context.Context, system string, messages []Message) (string, error)
	// Name — человекочитаемое имя провайдера для логов.
	Name() string
}
