package store

import (
	"os"
	"testing"
	"time"
)

// Тесты ходят в реальный локальный PostgreSQL (отдельная тестовая БД).
// Нет БД — тесты пропускаются, а не падают.
func newTestPostgres(t *testing.T) *Postgres {
	t.Helper()
	dsn := os.Getenv("GOPRACTICE_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:root@localhost:5432/gopractice_test"
	}
	p, err := NewPostgres(dsn)
	if err != nil {
		t.Skipf("PostgreSQL недоступен: %v", err)
	}
	t.Cleanup(func() {
		p.db.Exec(`TRUNCATE users, sessions, level_progress, task_progress, chat_messages, user_achievements, notes CASCADE`)
		p.Close()
	})
	// Чистим и на входе: прошлые упавшие тесты могли оставить мусор.
	p.db.Exec(`TRUNCATE users, sessions, level_progress, task_progress, chat_messages, user_achievements, notes CASCADE`)
	return p
}

func TestUsersAndSessions(t *testing.T) {
	p := newTestPostgres(t)

	// Первый — админ, второй — user.
	admin, err := p.CreateUser("petra", "hash-1")
	if err != nil {
		t.Fatal(err)
	}
	if admin.Role != "admin" {
		t.Fatalf("первый пользователь: роль %q", admin.Role)
	}
	user, err := p.CreateUser("vasya", "hash-2")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != "user" {
		t.Fatalf("второй пользователь: роль %q", user.Role)
	}
	if _, err := p.CreateUser("petra", "hash-3"); err != ErrUsernameTaken {
		t.Fatalf("дубликат имени: %v", err)
	}

	// Сессии: создание, чтение, продление, удаление, чистка протухших.
	if err := p.CreateSession("tok-1", admin.ID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	got, expires, err := p.UserBySession("tok-1")
	if err != nil || got == nil || got.ID != admin.ID {
		t.Fatalf("UserBySession: %v %v", got, err)
	}
	if err := p.TouchSession("tok-1", expires.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := p.CreateSession("tok-old", admin.ID, time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := p.UserBySession("tok-old"); got != nil {
		t.Fatal("протухшая сессия не должна резолвиться")
	}
	if err := p.DeleteExpiredSessions(); err != nil {
		t.Fatal(err)
	}
	if err := p.DeleteSession("tok-1"); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := p.UserBySession("tok-1"); got != nil {
		t.Fatal("удалённая сессия резолвится")
	}
}

func TestPerUserProgress(t *testing.T) {
	p := newTestPostgres(t)
	alice, _ := p.CreateUser("alice", "h")
	bob, _ := p.CreateUser("bob", "h")

	if err := p.RecordQuizAttempt(alice.ID, "c1", "l1", false); err != nil {
		t.Fatal(err)
	}
	if err := p.RecordQuizAttempt(alice.ID, "c1", "l1", true); err != nil {
		t.Fatal(err)
	}
	if err := p.RecordQuizAttempt(alice.ID, "c1", "l1", false); err != nil {
		t.Fatal(err)
	}
	if err := p.SaveDraft(alice.ID, "c1", "l1", "t1", "draft-1"); err != nil {
		t.Fatal(err)
	}
	if err := p.RecordTaskRun(alice.ID, "c1", "l1", "t1", "draft-2", true); err != nil {
		t.Fatal(err)
	}

	facts, err := p.CourseFacts(alice.ID, "c1")
	if err != nil {
		t.Fatal(err)
	}
	if lf := facts.Level("l1"); !lf.QuizPassed || lf.QuizAttempts != 3 {
		t.Fatalf("level facts: %+v", lf)
	}
	if tf := facts.Task("l1", "t1"); !tf.Completed || tf.Attempts != 1 || tf.Draft != "draft-2" || tf.CompletedAt.IsZero() {
		t.Fatalf("task facts: %+v", tf)
	}

	// Прогресс Боба пуст — факты не перетекают между пользователями.
	bobFacts, err := p.CourseFacts(bob.ID, "c1")
	if err != nil {
		t.Fatal(err)
	}
	if len(bobFacts.Levels) != 0 || len(bobFacts.Tasks) != 0 {
		t.Fatal("прогресс утёк между пользователями")
	}
}

func TestChatHistoryAndPrune(t *testing.T) {
	p := newTestPostgres(t)
	u, _ := p.CreateUser("chatty", "h")

	for i := 0; i < 205; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		if err := p.AppendChat(u.ID, "c1", "l1", role, "msg"); err != nil {
			t.Fatal(err)
		}
	}
	all, err := p.ChatHistory(u.ID, "c1", "l1", 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != chatKeepRows {
		t.Fatalf("прунинг не сработал: %d строк", len(all))
	}
	last10, _ := p.ChatHistory(u.ID, "c1", "l1", 10)
	if len(last10) != 10 {
		t.Fatalf("limit не работает: %d", len(last10))
	}
	if err := p.ClearChat(u.ID, "c1", "l1"); err != nil {
		t.Fatal(err)
	}
	if cleared, _ := p.ChatHistory(u.ID, "c1", "l1", 10); len(cleared) != 0 {
		t.Fatal("чат не очистился")
	}
}

func TestNotes(t *testing.T) {
	p := newTestPostgres(t)
	alice, _ := p.CreateUser("alice", "h")
	bob, _ := p.CreateUser("bob", "h")

	// Нет заметки — (nil, nil).
	if n, err := p.Note(alice.ID, "go", "l1"); err != nil || n != nil {
		t.Fatalf("пустая заметка: %v %v", n, err)
	}

	// Сохранение и апсерт.
	if err := p.SaveNote(alice.ID, "go", "l1", "первая версия"); err != nil {
		t.Fatal(err)
	}
	if err := p.SaveNote(alice.ID, "go", "l1", "вторая версия"); err != nil {
		t.Fatal(err)
	}
	if err := p.SaveNote(alice.ID, "go", "l2", "другой уровень"); err != nil {
		t.Fatal(err)
	}
	n, err := p.Note(alice.ID, "go", "l1")
	if err != nil || n == nil || n.Body != "вторая версия" || n.UpdatedAt.IsZero() {
		t.Fatalf("заметка после апсерта: %+v %v", n, err)
	}

	// Список: только свои, свежие сверху.
	if err := p.SaveNote(bob.ID, "go", "l1", "чужая"); err != nil {
		t.Fatal(err)
	}
	all, err := p.NotesAll(alice.ID)
	if err != nil || len(all) != 2 {
		t.Fatalf("NotesAll: %v %v", all, err)
	}
	for _, note := range all {
		if note.Body == "чужая" {
			t.Fatal("заметка утекла между пользователями")
		}
	}

	// Пустое тело удаляет заметку.
	if err := p.SaveNote(alice.ID, "go", "l1", "  \n "); err != nil {
		t.Fatal(err)
	}
	if n, _ := p.Note(alice.ID, "go", "l1"); n != nil {
		t.Fatal("пустое тело должно удалять заметку")
	}
	if all, _ := p.NotesAll(alice.ID); len(all) != 1 {
		t.Fatalf("после удаления должна остаться одна заметка, есть %d", len(all))
	}
}

func TestAchievements(t *testing.T) {
	p := newTestPostgres(t)
	u, _ := p.CreateUser("hero", "h")

	fresh, err := p.GrantAchievement(u.ID, "first-steps")
	if err != nil || !fresh {
		t.Fatalf("первая выдача: fresh=%v err=%v", fresh, err)
	}
	again, err := p.GrantAchievement(u.ID, "first-steps")
	if err != nil || again {
		t.Fatalf("повторная выдача должна быть no-op: fresh=%v err=%v", again, err)
	}
	all, err := p.Achievements(u.ID)
	if err != nil || len(all) != 1 {
		t.Fatalf("achievements: %v %v", all, err)
	}
}
