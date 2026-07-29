package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultModel подходит для частых коротких вопросов наставнику.
const DefaultModel = "gpt-5.6-luna"

const responsesEndpoint = "https://api.openai.com/v1/responses"

// APIProvider — OpenAI Responses API.
type APIProvider struct {
	client   *http.Client
	endpoint string
	apiKey   string
	model    string
}

// APIKeyConfigured сообщает, задан ли ключ OpenAI в окружении.
func APIKeyConfigured() bool {
	return strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != ""
}

func NewAPIProvider(model string) *APIProvider {
	if model == "" {
		model = DefaultModel
	}
	return &APIProvider{
		client:   &http.Client{Timeout: 90 * time.Second},
		endpoint: responsesEndpoint,
		apiKey:   os.Getenv("OPENAI_API_KEY"),
		model:    model,
	}
}

func (p *APIProvider) Name() string { return "api" }

func (p *APIProvider) Reply(ctx context.Context, system string, messages []Message) (string, error) {
	input := make([]responseInput, 0, len(messages))
	for _, m := range messages {
		role := "user"
		if m.Role == "assistant" {
			role = "assistant"
		}
		input = append(input, responseInput{Role: role, Content: m.Content})
	}

	payload, err := json.Marshal(responseRequest{
		Model:        p.model,
		Instructions: system,
		Input:        input,
		Store:        false,
		Reasoning:    responseReasoning{Effort: "low"},
	})
	if err != nil {
		return "", fmt.Errorf("кодирование запроса OpenAI: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("создание запроса OpenAI: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("запрос к OpenAI: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("чтение ответа OpenAI: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", openAIError(resp.StatusCode, body)
	}

	var parsed responseResult
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("разбор ответа OpenAI: %w", err)
	}
	var text strings.Builder
	for _, item := range parsed.Output {
		for _, content := range item.Content {
			if content.Type == "output_text" {
				text.WriteString(content.Text)
			}
		}
	}
	if text.Len() == 0 {
		return "", fmt.Errorf("пустой ответ модели OpenAI")
	}
	return text.String(), nil
}

type responseInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseReasoning struct {
	Effort string `json:"effort"`
}

type responseRequest struct {
	Model        string            `json:"model"`
	Instructions string            `json:"instructions"`
	Input        []responseInput   `json:"input"`
	Store        bool              `json:"store"`
	Reasoning    responseReasoning `json:"reasoning"`
}

type responseResult struct {
	Output []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func openAIError(status int, body []byte) error {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("чат не настроен: проверьте OPENAI_API_KEY")
	case http.StatusTooManyRequests:
		return fmt.Errorf("наставник перегружен, попробуйте чуть позже")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return fmt.Errorf("OpenAI временно недоступен, попробуйте чуть позже")
	}
	var apiErr struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
		return fmt.Errorf("OpenAI API: %s", apiErr.Error.Message)
	}
	return fmt.Errorf("OpenAI API вернул статус %d", status)
}
