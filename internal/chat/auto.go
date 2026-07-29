package chat

import "log"

// Select выбирает провайдера по режиму:
//   - "api"  — только Anthropic API (нужен ключ в окружении)
//   - "cli"  — только Claude Code CLI (подписка)
//   - "off"  — чат выключен
//   - "auto" — ключ есть → api; CLI установлен → cli; иначе выключен
//
// Возвращает nil, если чат недоступен.
func Select(mode, model string) Provider {
	switch mode {
	case "off":
		return nil
	case "api":
		if !APIKeyConfigured() {
			log.Print("чат: режим api, но ANTHROPIC_API_KEY не задан — чат выключен")
			return nil
		}
		return NewAPIProvider(model)
	case "cli":
		bin := FindCLI()
		if bin == "" {
			log.Print("чат: режим cli, но claude не найден — чат выключен")
			return nil
		}
		log.Printf("чат: Claude Code CLI (%s)", bin)
		return NewCLIProvider(bin, model)
	default: // auto
		if APIKeyConfigured() {
			log.Print("чат: Anthropic API (ключ из окружения)")
			return NewAPIProvider(model)
		}
		if bin := FindCLI(); bin != "" {
			log.Printf("чат: Claude Code CLI на подписке (%s)", bin)
			return NewCLIProvider(bin, model)
		}
		log.Print("чат: провайдер не найден (нет ключа API и claude CLI) — доступны только офлайн-подсказки")
		return nil
	}
}
