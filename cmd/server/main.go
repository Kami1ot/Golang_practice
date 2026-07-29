package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"gopractice/internal/chat"
	"gopractice/internal/content"
	"gopractice/internal/httpapi"
	"gopractice/internal/runner"
	"gopractice/internal/store"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080",
		"адрес HTTP-сервера (для доступа по локальной сети: 0.0.0.0:8080)")
	contentDir := flag.String("content", "content", "папка с курсами")
	webDir := flag.String("web", "web", "папка со статикой фронтенда")
	dsn := flag.String("dsn", envOr("GOPRACTICE_DSN", "postgres://postgres:root@localhost:5432/gopractice"),
		"PostgreSQL DSN; пустая строка — хранить всё в памяти (пропадёт при выходе)")
	chatMode := flag.String("chat", envOr("GOPRACTICE_CHAT", "auto"),
		"чат-наставник: auto|api|cli|off (api — ключ ANTHROPIC_API_KEY, cli — Claude Code на подписке)")
	chatModel := flag.String("chat-model", os.Getenv("GOPRACTICE_CHAT_MODEL"),
		"модель для чата (по умолчанию — claude-opus-4-8 для api, модель подписки для cli)")
	flag.Parse()

	catalog, err := content.Load(*contentDir)
	if err != nil {
		log.Fatalf("ошибка загрузки контента:\n%v", err)
	}
	log.Printf("контент загружен: курсов — %d", len(catalog.Order))
	httpapi.SweepTrash(*contentDir)

	var st store.Store
	if *dsn == "" {
		log.Print("ВНИМАНИЕ: -dsn пуст — пользователи, прогресс и чаты хранятся в памяти и пропадут при выходе")
		st = store.NewMemory()
	} else {
		st, err = store.NewPostgres(*dsn)
		if err != nil {
			log.Fatalf("PostgreSQL: %v\nПодсказка: проверьте, запущена ли служба postgresql-x64-17", err)
		}
	}
	defer st.Close()
	if err := st.DeleteExpiredSessions(); err != nil {
		log.Printf("чистка сессий: %v", err)
	}

	rn, err := runner.New()
	if err != nil {
		log.Fatalf("runner: %v", err)
	}
	rn.PreWarm()

	api := httpapi.New(catalog, st, rn, chat.Select(*chatMode, *chatModel), *contentDir, *webDir)

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Printf("не удалось открыть %s: %v", *addr, err)
		os.Exit(1)
	}
	log.Printf("GoPractice запущен: http://%s", *addr)
	if err := http.Serve(ln, api.Handler()); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
