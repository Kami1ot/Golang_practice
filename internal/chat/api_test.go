package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIProviderReplyUsesResponsesAPI(t *testing.T) {
	var got responseRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gotAuth := r.Header.Get("Authorization"); gotAuth != "Bearer test-key" {
			t.Errorf("Authorization = %q", gotAuth)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"output":[{"type":"reasoning"},{"type":"message","content":[{"type":"output_text","text":"Первый "},{"type":"output_text","text":"ответ"}]}]}`))
	}))
	defer server.Close()

	p := &APIProvider{client: server.Client(), endpoint: server.URL, apiKey: "test-key", model: "test-model"}
	gotReply, err := p.Reply(context.Background(), "Системные правила", []Message{
		{Role: "user", Content: "Вопрос"},
		{Role: "assistant", Content: "Предыдущий ответ"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotReply != "Первый ответ" {
		t.Fatalf("reply = %q", gotReply)
	}
	if got.Model != "test-model" || got.Instructions != "Системные правила" {
		t.Fatalf("request = %+v", got)
	}
	if got.Store {
		t.Fatal("request must disable storage")
	}
	if got.Reasoning.Effort != "low" {
		t.Fatalf("reasoning effort = %q", got.Reasoning.Effort)
	}
	if len(got.Input) != 2 || got.Input[1].Role != "assistant" {
		t.Fatalf("input = %+v", got.Input)
	}
}

func TestOpenAIError(t *testing.T) {
	err := openAIError(http.StatusTooManyRequests, nil)
	if err == nil || err.Error() != "наставник перегружен, попробуйте чуть позже" {
		t.Fatalf("error = %v", err)
	}
}
