package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopractice/internal/chat"
	"gopractice/internal/content"
	"gopractice/internal/store"
)

// ── Инфраструктура ──────────────────────────────────────────────────────────

func writeFixtureTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func fixtureFiles() map[string]string {
	return map[string]string{
		"courses/demo/course.json": `{"id":"demo","title":"Демо","description":"d","tags":["демо","го"],"levels":["l1","l2"]}`,
		"courses/demo/levels/l1/level.json": `{"id":"l1","title":"Первый","summary":"s","requires":[],"xp":10,"group":"Основы"}`,
		"courses/demo/levels/l1/lesson.md":  "# Урок\nСекретное слово урока: лягушка.",
		"courses/demo/levels/l1/quiz.json": `{"passScore":1,"questions":[
			{"id":"q1","type":"single","prompt":"?","options":["a","b"],"answer":1,"explanation":"объяснение-1"},
			{"id":"q2","type":"blank","prompt":"?","answers":["var"],"explanation":"объяснение-2"},
			{"id":"q3","type":"output","prompt":"?","code":"x","answer":"7\n","explanation":"объяснение-3"},
			{"id":"q4","type":"multi","prompt":"?","options":["a","b","c"],"answers":[0,2],"explanation":"объяснение-4"}]}`,
		"courses/demo/levels/l1/tasks/t1/task.json": `{"id":"t1","type":"unittest","title":"Задача","order":1,
			"hints":["подсказка-один","подсказка-два"]}`,
		"courses/demo/levels/l1/tasks/t1/statement.md":  "Реализуйте функцию Sum.",
		"courses/demo/levels/l1/tasks/t1/starter.go":    "package main\nfunc Sum(a, b int) int { return 0 }\n",
		"courses/demo/levels/l1/tasks/t1/main_test.go":  "package main\n// СЕКРЕТНЫЙ-ТЕСТ-МАРКЕР\nimport \"testing\"\nfunc TestSum(t *testing.T) { if Sum(1,2) != 3 { t.Fail() } }\n",
		"courses/demo/levels/l2/level.json": `{"id":"l2","title":"Второй","summary":"s","requires":["l1"],"xp":10}`,
		"courses/demo/levels/l2/lesson.md":  "# Урок 2",
	}
}

type testEnv struct {
	t    *testing.T
	api  *API
	srv  *httptest.Server
	root string
}

func newTestEnv(t *testing.T, ch chat.Provider) *testEnv {
	t.Helper()
	root := t.TempDir()
	writeFixtureTree(t, root, fixtureFiles())
	cat, err := content.Load(root)
	if err != nil {
		t.Fatalf("фикстура не загрузилась: %v", err)
	}
	api := New(cat, store.NewMemory(), nil, ch, root, t.TempDir())
	srv := httptest.NewServer(api.Handler())
	t.Cleanup(srv.Close)
	return &testEnv{t: t, api: api, srv: srv, root: root}
}

// client — HTTP-клиент с cookie jar (после register/login несёт сессию).
func (e *testEnv) client() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

func (e *testEnv) register(cl *http.Client, username string) userDTO {
	e.t.Helper()
	resp, body := e.request(cl, "POST", "/api/auth/register",
		fmt.Sprintf(`{"username":%q,"password":"password123"}`, username))
	if resp.StatusCode != http.StatusCreated {
		e.t.Fatalf("регистрация %s: код %d: %s", username, resp.StatusCode, body)
	}
	var u userDTO
	json.Unmarshal([]byte(body), &u)
	return u
}

func (e *testEnv) request(cl *http.Client, method, path, body string) (*http.Response, string) {
	e.t.Helper()
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, e.srv.URL+path, rd)
	if err != nil {
		e.t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := cl.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		e.t.Fatal(err)
	}
	return resp, string(data)
}

// ── Авторизация ─────────────────────────────────────────────────────────────

func TestAuthFlow(t *testing.T) {
	e := newTestEnv(t, nil)

	// Без сессии API закрыт.
	resp, _ := e.request(e.client(), "GET", "/api/courses", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("без входа код %d, ожидался 401", resp.StatusCode)
	}

	// Первый зарегистрированный — админ.
	admin := e.client()
	u := e.register(admin, "petra")
	if u.Role != "admin" {
		t.Fatalf("первый пользователь должен быть админом, роль %q", u.Role)
	}
	resp, body := e.request(admin, "GET", "/api/auth/me", "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, `"petra"`) {
		t.Fatalf("me: %d %s", resp.StatusCode, body)
	}
	resp, _ = e.request(admin, "GET", "/api/courses", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("с сессией код %d", resp.StatusCode)
	}

	// Второй — обычный пользователь; занятое имя — 409.
	second := e.client()
	if u2 := e.register(second, "vasya"); u2.Role != "user" {
		t.Fatalf("второй пользователь должен быть user, роль %q", u2.Role)
	}
	resp, _ = e.request(e.client(), "POST", "/api/auth/register", `{"username":"petra","password":"password123"}`)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("повторное имя: код %d, ожидался 409", resp.StatusCode)
	}

	// Неверный пароль.
	resp, _ = e.request(e.client(), "POST", "/api/auth/login", `{"username":"petra","password":"не-тот"}`)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("неверный пароль: код %d", resp.StatusCode)
	}

	// Логин + логаут.
	cl := e.client()
	resp, _ = e.request(cl, "POST", "/api/auth/login", `{"username":"petra","password":"password123"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("логин: код %d", resp.StatusCode)
	}
	resp, _ = e.request(cl, "POST", "/api/auth/logout", "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("логаут: код %d", resp.StatusCode)
	}
	resp, _ = e.request(cl, "GET", "/api/courses", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("после логаута код %d, ожидался 401", resp.StatusCode)
	}
}

// ── Существующее поведение под авторизацией ────────────────────────────────

func TestLevelContentStripsAnswers(t *testing.T) {
	e := newTestEnv(t, nil)
	cl := e.client()
	e.register(cl, "student")

	resp, body := e.request(cl, "GET", "/api/courses/demo/levels/l1", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("код %d", resp.StatusCode)
	}
	for _, forbidden := range []string{`"answer"`, `"answers"`, "объяснение-", "СЕКРЕТНЫЙ-ТЕСТ-МАРКЕР"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("ответ уровня раскрывает %s", forbidden)
		}
	}
	if !strings.Contains(body, "подсказка-один") {
		t.Fatal("hints не дошли до клиента")
	}
}

func TestLockedLevel(t *testing.T) {
	e := newTestEnv(t, nil)
	cl := e.client()
	e.register(cl, "student")
	resp, _ := e.request(cl, "GET", "/api/courses/demo/levels/l2", "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("код %d, ожидался 403", resp.StatusCode)
	}
}

func TestQuizGradingAndPerUserProgress(t *testing.T) {
	e := newTestEnv(t, nil)
	alice := e.client()
	e.register(alice, "alice")

	// Неправильные ответы: passed=false, объяснения приходят.
	resp, body := e.request(alice, "POST", "/api/courses/demo/levels/l1/quiz",
		`{"answers":{"q1":0,"q2":"const","q3":"8","q4":[0,1]}}`)
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, `"passed":false`) {
		t.Fatalf("нefail: %d %s", resp.StatusCode, body)
	}

	// Правильные.
	_, body = e.request(alice, "POST", "/api/courses/demo/levels/l1/quiz",
		`{"answers":{"q1":1,"q2":" var ","q3":"7","q4":[2,0]}}`)
	if !strings.Contains(body, `"passed":true`) {
		t.Fatalf("правильные ответы не засчитаны: %s", body)
	}

	// Прогресс личный: у Боба тест не сдан.
	bob := e.client()
	e.register(bob, "bob")
	_, body = e.request(bob, "GET", "/api/courses/demo/tree", "")
	if !strings.Contains(body, `"quizPassed":false`) {
		t.Fatalf("прогресс утёк между пользователями: %s", body)
	}
	_, body = e.request(alice, "GET", "/api/courses/demo/tree", "")
	if !strings.Contains(body, `"quizPassed":true`) {
		t.Fatalf("прогресс Алисы потерян: %s", body)
	}
	if !strings.Contains(body, `"group":"Основы"`) {
		t.Fatalf("группа уровня не дошла до дерева: %s", body)
	}
}

func TestDraftRoundtrip(t *testing.T) {
	e := newTestEnv(t, nil)
	cl := e.client()
	e.register(cl, "student")
	resp, _ := e.request(cl, "PUT", "/api/courses/demo/levels/l1/tasks/t1/draft",
		`{"code":"package main // мой черновик"}`)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("код %d", resp.StatusCode)
	}
	_, body := e.request(cl, "GET", "/api/courses/demo/levels/l1", "")
	if !strings.Contains(body, "мой черновик") {
		t.Fatal("черновик не вернулся")
	}
}

func TestNotesRoundtrip(t *testing.T) {
	e := newTestEnv(t, nil)
	cl := e.client()
	e.register(cl, "student")

	// До сохранения — пустая заметка.
	resp, body := e.request(cl, "GET", "/api/courses/demo/levels/l1/note", "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, `"body":""`) {
		t.Fatalf("пустая заметка: %d %s", resp.StatusCode, body)
	}

	// Сохранение и чтение.
	resp, _ = e.request(cl, "PUT", "/api/courses/demo/levels/l1/note", `{"body":"## моя заметка"}`)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("PUT note: код %d", resp.StatusCode)
	}
	_, body = e.request(cl, "GET", "/api/courses/demo/levels/l1/note", "")
	if !strings.Contains(body, "моя заметка") {
		t.Fatalf("заметка не вернулась: %s", body)
	}

	// Список для кабинета: имя собирается из каталога (курс/группа/уровень).
	_, body = e.request(cl, "GET", "/api/notes", "")
	for _, want := range []string{`"courseTitle":"Демо"`, `"group":"Основы"`, `"levelTitle":"Первый"`, "моя заметка"} {
		if !strings.Contains(body, want) {
			t.Fatalf("в списке заметок нет %s: %s", want, body)
		}
	}

	// Заметка закрытого уровня недоступна.
	resp, _ = e.request(cl, "PUT", "/api/courses/demo/levels/l2/note", `{"body":"x"}`)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("заметка закрытого уровня: код %d", resp.StatusCode)
	}

	// Пустое тело удаляет заметку из списка.
	e.request(cl, "PUT", "/api/courses/demo/levels/l1/note", `{"body":""}`)
	_, body = e.request(cl, "GET", "/api/notes", "")
	if body != "[]\n" && strings.Contains(body, "моя заметка") {
		t.Fatalf("заметка не удалилась: %s", body)
	}

	// Чужому пользователю заметки не видны.
	other := e.client()
	e.register(other, "other")
	e.request(cl, "PUT", "/api/courses/demo/levels/l1/note", `{"body":"секрет"}`)
	_, body = e.request(other, "GET", "/api/notes", "")
	if strings.Contains(body, "секрет") {
		t.Fatal("заметка утекла между пользователями")
	}
}

func TestContentWhitelist(t *testing.T) {
	e := newTestEnv(t, nil)
	cl := e.client()
	e.register(cl, "student")
	resp, _ := e.request(cl, "GET", "/content/courses/demo/levels/l1/quiz.json", "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("quiz.json должен быть недоступен, код %d", resp.StatusCode)
	}
	resp, _ = e.request(cl, "GET", "/content/courses/demo/levels/l1/tasks/t1/main_test.go", "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("main_test.go должен быть недоступен, код %d", resp.StatusCode)
	}
}

// ── Админ-редактор ──────────────────────────────────────────────────────────

func TestAdminRequiresRole(t *testing.T) {
	e := newTestEnv(t, nil)
	admin := e.client()
	e.register(admin, "root") // первый = админ
	user := e.client()
	e.register(user, "plain")

	resp, _ := e.request(user, "POST", "/api/admin/courses", `{"id":"x","title":"X"}`)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("не-админ создал курс: код %d", resp.StatusCode)
	}
	resp, _ = e.request(user, "POST", "/api/reload", "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("не-админ перезагрузил контент: код %d", resp.StatusCode)
	}
	resp, _ = e.request(user, "GET", "/api/admin/courses/demo/levels/l1", "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("не-админ прочитал ответы: код %d", resp.StatusCode)
	}
	// А админ ответы видит (это единственная легальная поверхность).
	resp, body := e.request(admin, "GET", "/api/admin/courses/demo/levels/l1", "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, "СЕКРЕТНЫЙ-ТЕСТ-МАРКЕР") {
		t.Fatalf("админ должен видеть main_test.go: %d", resp.StatusCode)
	}
}

func TestAdminCRUDAndRevert(t *testing.T) {
	e := newTestEnv(t, nil)
	admin := e.client()
	e.register(admin, "root")

	// Создание курса → уровня → задачи.
	resp, body := e.request(admin, "POST", "/api/admin/courses",
		`{"id":"new-course","title":"Новый","description":"тест"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("создание курса: %d %s", resp.StatusCode, body)
	}
	resp, body = e.request(admin, "POST", "/api/admin/courses/new-course/levels",
		`{"id":"lv1","title":"Уровень","summary":"s","requires":[],"xp":10}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("создание уровня: %d %s", resp.StatusCode, body)
	}
	resp, body = e.request(admin, "POST", "/api/admin/courses/new-course/levels/lv1/tasks",
		`{"id":"tk1","type":"io","title":"Задача","required":true,"order":1,
		  "tests":[{"name":"т","stdin":"","stdout":"ok\n"}],
		  "statementMd":"Сделай.","starterCode":"package main\nfunc main(){}\n"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("создание задачи: %d %s", resp.StatusCode, body)
	}

	// Новый контент виден в публичном API сразу (hot-swap).
	_, body = e.request(admin, "GET", "/api/courses", "")
	if !strings.Contains(body, "new-course") {
		t.Fatalf("новый курс не виден: %s", body)
	}

	// Битый quiz откатывается: 422 + диск байт-в-байт + каталог живой.
	quizPath := filepath.Join(e.root, "courses", "demo", "levels", "l1", "quiz.json")
	before, _ := os.ReadFile(quizPath)
	resp, body = e.request(admin, "PUT", "/api/admin/courses/demo/levels/l1/quiz",
		`{"passScore":1,"questions":[{"id":"q1","type":"странный","prompt":"?"}]}`)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("битый quiz: код %d %s", resp.StatusCode, body)
	}
	after, _ := os.ReadFile(quizPath)
	if string(before) != string(after) {
		t.Fatal("диск не откатился побайтово")
	}
	resp, _ = e.request(admin, "GET", "/api/courses/demo/levels/l1", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("каталог сломался после отката: %d", resp.StatusCode)
	}

	// Path traversal и зарезервированные имена.
	resp, _ = e.request(admin, "POST", "/api/admin/courses", `{"id":"../evil","title":"X"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("traversal-id прошёл: код %d", resp.StatusCode)
	}
	resp, _ = e.request(admin, "POST", "/api/admin/courses", `{"id":"con","title":"X"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("зарезервированное имя прошло: код %d", resp.StatusCode)
	}

	// Удаление уровня чистит course.json и папку.
	resp, _ = e.request(admin, "DELETE", "/api/admin/courses/new-course/levels/lv1", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("удаление уровня: %d", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(e.root, "courses", "new-course", "levels", "lv1")); !os.IsNotExist(err) {
		t.Fatal("папка уровня не удалена")
	}
}

// ── Чат ─────────────────────────────────────────────────────────────────────

// fakeProvider записывает системный промпт для проверок.
type fakeProvider struct {
	lastSystem string
	lastMsgs   []chat.Message
}

func (f *fakeProvider) Name() string { return "fake" }
func (f *fakeProvider) Reply(_ context.Context, system string, msgs []chat.Message) (string, error) {
	f.lastSystem = system
	f.lastMsgs = msgs
	return "Наводящий вопрос: а что говорит урок про сумму?", nil
}

func TestChatDisabled(t *testing.T) {
	e := newTestEnv(t, nil)
	cl := e.client()
	e.register(cl, "student")
	_, body := e.request(cl, "GET", "/api/courses/demo/levels/l1/chat", "")
	if !strings.Contains(body, `"enabled":false`) {
		t.Fatalf("чат должен быть выключен: %s", body)
	}
	resp, _ := e.request(cl, "POST", "/api/courses/demo/levels/l1/chat", `{"message":"привет"}`)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("POST в выключенный чат: код %d", resp.StatusCode)
	}
}

func TestChatPromptSafety(t *testing.T) {
	fake := &fakeProvider{}
	e := newTestEnv(t, fake)
	cl := e.client()
	e.register(cl, "student")

	resp, body := e.request(cl, "POST", "/api/courses/demo/levels/l1/chat",
		`{"message":"как решить задачу?","taskId":"t1"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("чат: %d %s", resp.StatusCode, body)
	}

	// В промпте есть то, что ученик и так видит...
	for _, want := range []string{"лягушка", "Реализуйте функцию Sum", "func Sum(a, b int) int { return 0 }"} {
		if !strings.Contains(fake.lastSystem, want) {
			t.Fatalf("в промпте нет %q", want)
		}
	}
	// ...и НИКОГДА — ответов и скрытых тестов.
	for _, forbidden := range []string{"СЕКРЕТНЫЙ-ТЕСТ-МАРКЕР", "объяснение-1", `"answer"`} {
		if strings.Contains(fake.lastSystem, forbidden) {
			t.Fatalf("ПРОМПТ РАСКРЫВАЕТ %q", forbidden)
		}
	}

	// История хранится на сервере и попадает в контекст следующего запроса.
	e.request(cl, "POST", "/api/courses/demo/levels/l1/chat", `{"message":"а подробнее?"}`)
	if len(fake.lastMsgs) < 3 {
		t.Fatalf("история не попала в контекст: %d сообщений", len(fake.lastMsgs))
	}
	// Чат закрытого уровня недоступен.
	resp, _ = e.request(cl, "POST", "/api/courses/demo/levels/l2/chat", `{"message":"пусти"}`)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("чат закрытого уровня: код %d", resp.StatusCode)
	}
}

// ── Ачивки и кабинет ────────────────────────────────────────────────────────

func TestProfileAndAchievements(t *testing.T) {
	e := newTestEnv(t, nil)
	cl := e.client()
	u := e.register(cl, "student")

	// До прогресса: ачивок нет, всё залочено.
	_, body := e.request(cl, "GET", "/api/profile", "")
	if strings.Contains(body, `"grantedAt":"2`) {
		t.Fatalf("ачивки выданы авансом: %s", body)
	}

	// Собираем факты напрямую в store: тест с 1-й попытки + задача решена.
	st := e.api.store
	st.RecordQuizAttempt(u.ID, "demo", "l1", true)
	st.RecordTaskRun(u.ID, "demo", "l1", "t1", "code", true)

	_, body = e.request(cl, "GET", "/api/profile", "")
	for _, want := range []string{"first-steps", "quiz-perfect", "task-first-try", "course-demo"} {
		if !strings.Contains(body, `"id":"`+want+`"`) {
			t.Fatalf("нет ачивки %s: %s", want, body)
		}
	}
	var profile profileDTO
	json.Unmarshal([]byte(body), &profile)
	granted := 0
	for _, a := range profile.Achievements {
		if a.GrantedAt != nil {
			granted++
		}
	}
	// first-steps, quiz-perfect, task-first-try, course-demo (l1+l2 транзитивно).
	if granted != 4 {
		t.Fatalf("выдано %d ачивок, ожидалось 4: %s", granted, body)
	}
	if profile.Stats.XP != 20 || profile.Stats.LevelsCompleted != 2 {
		t.Fatalf("статистика: %+v", profile.Stats)
	}

	// Повторный вызов не дублирует выдачу.
	_, body2 := e.request(cl, "GET", "/api/profile", "")
	var profile2 profileDTO
	json.Unmarshal([]byte(body2), &profile2)
	for i, a := range profile2.Achievements {
		if a.GrantedAt != nil && profile.Achievements[i].GrantedAt != nil &&
			!a.GrantedAt.Equal(*profile.Achievements[i].GrantedAt) {
			t.Fatal("granted_at изменился при повторной оценке")
		}
	}
}

// ── Витрина: теги и агрегаты по пользователям ───────────────────────────────

func TestCourseAggregatesAndTags(t *testing.T) {
	e := newTestEnv(t, nil)
	admin := e.client()
	uA := e.register(admin, "anna") // первый = админ (нужен для /api/reload)
	clB := e.client()
	uB := e.register(clB, "boris")

	// Сеем прогресс ДО первого GET /api/courses — агрегаты кэшируются с TTL.
	st := e.api.store
	st.RecordQuizAttempt(uA.ID, "demo", "l1", true)
	st.RecordTaskRun(uA.ID, "demo", "l1", "t1", "code", true) // курс завершён (l2 транзитивно)
	st.SaveDraft(uB.ID, "demo", "l1", "t1", "draft")          // начал, не завершил

	resp, body := e.request(admin, "GET", "/api/courses", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("courses: код %d: %s", resp.StatusCode, body)
	}
	var courses []courseDTO
	json.Unmarshal([]byte(body), &courses)
	if len(courses) != 1 {
		t.Fatalf("ожидался один курс: %s", body)
	}
	c := courses[0]
	if len(c.Tags) != 2 || c.Tags[0] != "демо" || c.Tags[1] != "го" {
		t.Fatalf("tags не доехали до DTO: %v", c.Tags)
	}
	if c.UsersCompleted != 1 || c.UsersInProgress != 1 {
		t.Fatalf("агрегаты: completed=%d inProgress=%d, ожидалось 1/1", c.UsersCompleted, c.UsersInProgress)
	}

	// Борис завершает курс; кэш ещё держит старые числа…
	st.RecordQuizAttempt(uB.ID, "demo", "l1", true)
	st.RecordTaskRun(uB.ID, "demo", "l1", "t1", "code", true)
	_, body = e.request(admin, "GET", "/api/courses", "")
	json.Unmarshal([]byte(body), &courses)
	if courses[0].UsersCompleted != 1 {
		t.Fatalf("кэш агрегатов не сработал: %s", body)
	}

	// …а reload сбрасывает кэш.
	if resp, body := e.request(admin, "POST", "/api/reload", ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("reload: %d %s", resp.StatusCode, body)
	}
	_, body = e.request(admin, "GET", "/api/courses", "")
	json.Unmarshal([]byte(body), &courses)
	if courses[0].UsersCompleted != 2 || courses[0].UsersInProgress != 0 {
		t.Fatalf("после reload агрегаты не пересчитались: %s", body)
	}
}

// Админ-обновление меты курса не должно терять теги (главная регрессия tags).
func TestAdminCourseUpdateKeepsTags(t *testing.T) {
	e := newTestEnv(t, nil)
	admin := e.client()
	e.register(admin, "root")

	// Обновление меты БЕЗ поля tags — теги из course.json сохраняются.
	resp, body := e.request(admin, "PUT", "/api/admin/courses/demo",
		`{"id":"demo","title":"Демо 2","description":"d2","levels":["l1","l2"]}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: %d %s", resp.StatusCode, body)
	}
	_, body = e.request(admin, "GET", "/api/admin/courses/demo", "")
	if !strings.Contains(body, `"демо"`) || !strings.Contains(body, `"го"`) {
		t.Fatalf("теги потеряны при обновлении меты: %s", body)
	}

	// Явная передача tags — заменяет список.
	resp, body = e.request(admin, "PUT", "/api/admin/courses/demo",
		`{"id":"demo","title":"Демо 2","description":"d2","tags":["новый"],"levels":["l1","l2"]}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update tags: %d %s", resp.StatusCode, body)
	}
	_, body = e.request(admin, "GET", "/api/admin/courses/demo", "")
	if !strings.Contains(body, `"новый"`) || strings.Contains(body, `"демо"`) {
		t.Fatalf("теги не заменились: %s", body)
	}
}

// ── Юнит: нормализация ответов quiz (без HTTP) ─────────────────────────────

func TestQuizNormalization(t *testing.T) {
	e := newTestEnv(t, nil)
	l1 := e.api.cat().Courses["demo"].Levels["l1"]

	q3 := l1.Quiz.Questions[2]
	if !gradeQuestion(q3, json.RawMessage(`"7\r\n"`)) {
		t.Fatal("output-ответ с CRLF должен засчитываться")
	}
	q4 := l1.Quiz.Questions[3]
	if !gradeQuestion(q4, json.RawMessage(`[2,0]`)) {
		t.Fatal("multi в другом порядке должен засчитываться")
	}
	if gradeQuestion(q4, json.RawMessage(`[0]`)) || gradeQuestion(q4, json.RawMessage(`[0,1,2]`)) {
		t.Fatal("multi с неполным/лишним набором не должен засчитываться")
	}
}
