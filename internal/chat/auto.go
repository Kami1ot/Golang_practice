package chat

import "log"

// Select выбирает провайдера по режиму:
//   - "api" и "auto" — OpenAI Responses API с ключом OPENAI_API_KEY
//   - "off" — чат выключен
//
// Возвращает nil, если чат недоступен.
func Select(mode, model string) Provider {
	switch mode {
	case "off":
		return nil
	case "api", "auto":
		if !APIKeyConfigured() {
			log.Print("чат: OPENAI_API_KEY не задан — чат выключен")
			return nil
		}
		log.Printf("чат: OpenAI Responses API (%s)", model)
		return NewAPIProvider(model)
	default:
		log.Printf("чат: неизвестный режим %q — чат выключен", mode)
		return nil
	}
}
