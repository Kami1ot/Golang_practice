package chat

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// APIProvider — Anthropic Messages API через официальный SDK.
type APIProvider struct {
	client anthropic.Client
	model  anthropic.Model
}

// APIKeyConfigured — задан ли ключ/токен в окружении.
func APIKeyConfigured() bool {
	return os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("ANTHROPIC_AUTH_TOKEN") != ""
}

func NewAPIProvider(model string, opts ...option.RequestOption) *APIProvider {
	if model == "" {
		model = string(anthropic.ModelClaudeOpus4_8)
	}
	return &APIProvider{
		client: anthropic.NewClient(opts...),
		model:  anthropic.Model(model),
	}
}

func (p *APIProvider) Name() string { return "api" }

func (p *APIProvider) Reply(ctx context.Context, system string, messages []Message) (string, error) {
	params := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: 2000,
		System:    []anthropic.TextBlockParam{{Text: system}},
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{},
		},
	}
	for _, m := range messages {
		if m.Role == "assistant" {
			params.Messages = append(params.Messages,
				anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		} else {
			params.Messages = append(params.Messages,
				anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		}
	}

	resp, err := p.client.Messages.New(ctx, params)
	if err != nil {
		var apierr *anthropic.Error
		if errors.As(err, &apierr) {
			switch apierr.StatusCode {
			case 401, 403:
				return "", fmt.Errorf("чат не настроен: проверьте ключ API")
			case 429, 529:
				return "", fmt.Errorf("наставник перегружен, попробуйте чуть позже")
			}
		}
		return "", err
	}
	if resp.StopReason == anthropic.StopReasonRefusal {
		return "Я не могу ответить на этот запрос. Спросите что-нибудь по теме урока!", nil
	}

	var out string
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			out += t.Text
		}
	}
	if out == "" {
		return "", fmt.Errorf("пустой ответ модели (stop_reason: %s)", resp.StopReason)
	}
	return out, nil
}
