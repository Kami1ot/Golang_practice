// Package httpapi — REST API и раздача статики.
package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"gopractice/internal/chat"
	"gopractice/internal/content"
	"gopractice/internal/runner"
	"gopractice/internal/store"
)

type API struct {
	mu         sync.RWMutex // защищает catalog (подменяется при reload)
	fsMu       sync.Mutex   // сериализует мутации content/ и content.Load
	catalog    *content.Catalog
	store      store.Store
	runner     *runner.Runner
	chat       chat.Provider // nil — чат выключен
	contentDir string
	webDir     string

	aggMu    sync.Mutex // защищает aggCache
	aggCache map[string]aggCacheEntry
}

func New(catalog *content.Catalog, st store.Store, rn *runner.Runner, ch chat.Provider, contentDir, webDir string) *API {
	return &API{
		catalog:    catalog,
		store:      st,
		runner:     rn,
		chat:       ch,
		contentDir: contentDir,
		webDir:     webDir,
		aggCache:   map[string]aggCacheEntry{},
	}
}

func (a *API) cat() *content.Catalog {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.catalog
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	// Открытые маршруты: health, вход/регистрация, статика фронтенда.
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	mux.HandleFunc("POST /api/auth/register", a.handleRegister)
	mux.HandleFunc("POST /api/auth/login", a.handleLogin)
	mux.HandleFunc("GET /api/auth/me", a.handleMe)
	mux.Handle("POST /api/auth/logout", a.requireUser(a.handleLogout))

	// Всё остальное API — только для вошедших.
	mux.Handle("GET /api/courses", a.requireUser(a.handleCourses))
	mux.Handle("GET /api/courses/{course}/tree", a.requireUser(a.handleTree))
	mux.Handle("GET /api/courses/{course}/levels/{level}", a.requireUser(a.handleLevel))
	mux.Handle("POST /api/courses/{course}/levels/{level}/quiz", a.requireUser(a.handleQuiz))
	mux.Handle("POST /api/courses/{course}/levels/{level}/tasks/{task}/run", a.requireUser(a.handleRun))
	mux.Handle("PUT /api/courses/{course}/levels/{level}/tasks/{task}/draft", a.requireUser(a.handleDraft))
	mux.Handle("GET /api/courses/{course}/levels/{level}/note", a.requireUser(a.handleNoteGet))
	mux.Handle("PUT /api/courses/{course}/levels/{level}/note", a.requireUser(a.handleNotePut))
	mux.Handle("GET /api/notes", a.requireUser(a.handleNotesList))
	mux.Handle("GET /api/courses/{course}/levels/{level}/chat", a.requireUser(a.handleChatHistory))
	mux.Handle("POST /api/courses/{course}/levels/{level}/chat", a.requireUser(a.handleChatSend))
	mux.Handle("DELETE /api/courses/{course}/levels/{level}/chat", a.requireUser(a.handleChatClear))
	mux.Handle("GET /api/profile", a.requireUser(a.handleProfile))

	// Перезагрузка контента и админ-редактор — только админ.
	mux.Handle("POST /api/reload", a.requireAdmin(a.handleReload))
	a.registerAdminRoutes(mux)

	// Картинки уроков: только whitelist расширений, чтобы из браузера
	// нельзя было скачать quiz.json с ответами или main_test.go.
	mux.Handle("GET /content/", http.StripPrefix("/content/",
		a.requireUser(func(w http.ResponseWriter, r *http.Request) {
			a.imagesOnly(http.FileServer(http.Dir(a.contentDir))).ServeHTTP(w, r)
		})))

	mux.Handle("GET /", noStaleCache(http.FileServer(http.Dir(a.webDir))))
	return mux
}

// noStaleCache заставляет браузер ревалидировать статику (иначе героическая
// эвристика кэша Chrome показывает старые css/js после правок файлов).
func noStaleCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

var imageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true, ".webp": true,
}

func (a *API) imagesOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !imageExts[strings.ToLower(filepath.Ext(r.URL.Path))] {
			writeErr(w, http.StatusForbidden, "доступны только изображения")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("кодирование ответа: %v", err)
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// readJSON читает тело запроса с лимитом 1 МБ.
func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeErr(w, http.StatusBadRequest, "некорректный JSON: "+err.Error())
		return false
	}
	return true
}
