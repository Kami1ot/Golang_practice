package store

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // драйвер database/sql "pgx"
)

//go:embed schema.sql
var schemaSQL string

type Postgres struct {
	db *sql.DB
}

var dbNameRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// NewPostgres подключается по DSN, при необходимости создаёт саму базу
// (через maintenance-БД postgres) и накатывает идемпотентную схему.
func NewPostgres(dsn string) (*Postgres, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("некорректный DSN: %w", err)
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	if !dbNameRe.MatchString(dbName) {
		return nil, fmt.Errorf("некорректное имя базы в DSN: %q", dbName)
	}

	if err := ensureDatabase(u, dbName); err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("подключение к %s: %w", dbName, err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("миграция схемы: %w", err)
	}
	return &Postgres{db: db}, nil
}

// ensureDatabase создаёт базу dbName, если её ещё нет.
func ensureDatabase(u *url.URL, dbName string) error {
	maint := *u
	maint.Path = "/postgres"
	db, err := sql.Open("pgx", maint.String())
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("подключение к PostgreSQL: %w", err)
	}
	var exists bool
	if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`, dbName).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		// Имя проверено регуляркой, инъекция исключена.
		if _, err := db.Exec(fmt.Sprintf(`CREATE DATABASE %q`, dbName)); err != nil {
			return fmt.Errorf("создание базы %s: %w", dbName, err)
		}
	}
	return nil
}

// ── Прогресс ────────────────────────────────────────────────────────────────

func (p *Postgres) CourseFacts(userID int64, courseID string) (*CourseFacts, error) {
	facts := &CourseFacts{
		Levels: map[string]LevelFacts{},
		Tasks:  map[string]map[string]TaskFacts{},
	}

	rows, err := p.db.Query(
		`SELECT level_id, quiz_passed, quiz_attempts
		   FROM level_progress WHERE user_id = $1 AND course_id = $2`, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var f LevelFacts
		if err := rows.Scan(&id, &f.QuizPassed, &f.QuizAttempts); err != nil {
			return nil, err
		}
		facts.Levels[id] = f
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	taskRows, err := p.db.Query(
		`SELECT level_id, task_id, completed, attempts, draft, completed_at
		   FROM task_progress WHERE user_id = $1 AND course_id = $2`, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer taskRows.Close()
	for taskRows.Next() {
		var levelID, taskID string
		var f TaskFacts
		var completedAt sql.NullTime
		if err := taskRows.Scan(&levelID, &taskID, &f.Completed, &f.Attempts, &f.Draft, &completedAt); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			f.CompletedAt = completedAt.Time
		}
		if facts.Tasks[levelID] == nil {
			facts.Tasks[levelID] = map[string]TaskFacts{}
		}
		facts.Tasks[levelID][taskID] = f
	}
	return facts, taskRows.Err()
}

func (p *Postgres) CourseFactsAll(courseID string) (map[int64]*CourseFacts, error) {
	byUser := map[int64]*CourseFacts{}
	userFacts := func(userID int64) *CourseFacts {
		f := byUser[userID]
		if f == nil {
			f = &CourseFacts{
				Levels: map[string]LevelFacts{},
				Tasks:  map[string]map[string]TaskFacts{},
			}
			byUser[userID] = f
		}
		return f
	}

	rows, err := p.db.Query(
		`SELECT user_id, level_id, quiz_passed, quiz_attempts
		   FROM level_progress WHERE course_id = $1 AND user_id IS NOT NULL`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID int64
		var id string
		var f LevelFacts
		if err := rows.Scan(&userID, &id, &f.QuizPassed, &f.QuizAttempts); err != nil {
			return nil, err
		}
		userFacts(userID).Levels[id] = f
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	taskRows, err := p.db.Query(
		`SELECT user_id, level_id, task_id, completed, attempts, draft, completed_at
		   FROM task_progress WHERE course_id = $1 AND user_id IS NOT NULL`, courseID)
	if err != nil {
		return nil, err
	}
	defer taskRows.Close()
	for taskRows.Next() {
		var userID int64
		var levelID, taskID string
		var f TaskFacts
		var completedAt sql.NullTime
		if err := taskRows.Scan(&userID, &levelID, &taskID, &f.Completed, &f.Attempts, &f.Draft, &completedAt); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			f.CompletedAt = completedAt.Time
		}
		facts := userFacts(userID)
		if facts.Tasks[levelID] == nil {
			facts.Tasks[levelID] = map[string]TaskFacts{}
		}
		facts.Tasks[levelID][taskID] = f
	}
	return byUser, taskRows.Err()
}

func (p *Postgres) RecordQuizAttempt(userID int64, courseID, levelID string, passed bool) error {
	_, err := p.db.Exec(`
		INSERT INTO level_progress (user_id, course_id, level_id, quiz_passed, quiz_attempts)
		VALUES ($1, $2, $3, $4, 1)
		ON CONFLICT (user_id, course_id, level_id) DO UPDATE SET
			quiz_attempts = level_progress.quiz_attempts + 1,
			quiz_passed   = level_progress.quiz_passed OR EXCLUDED.quiz_passed,
			updated_at    = now()`,
		userID, courseID, levelID, passed)
	return err
}

func (p *Postgres) RecordTaskRun(userID int64, courseID, levelID, taskID, draft string, passed bool) error {
	var completedAt *time.Time
	if passed {
		now := time.Now()
		completedAt = &now
	}
	_, err := p.db.Exec(`
		INSERT INTO task_progress (user_id, course_id, level_id, task_id, completed, attempts, draft, completed_at)
		VALUES ($1, $2, $3, $4, $5, 1, $6, $7)
		ON CONFLICT (user_id, course_id, level_id, task_id) DO UPDATE SET
			attempts     = task_progress.attempts + 1,
			draft        = EXCLUDED.draft,
			completed    = task_progress.completed OR EXCLUDED.completed,
			completed_at = COALESCE(task_progress.completed_at, EXCLUDED.completed_at),
			updated_at   = now()`,
		userID, courseID, levelID, taskID, passed, draft, completedAt)
	return err
}

func (p *Postgres) SaveDraft(userID int64, courseID, levelID, taskID, draft string) error {
	_, err := p.db.Exec(`
		INSERT INTO task_progress (user_id, course_id, level_id, task_id, draft)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, course_id, level_id, task_id) DO UPDATE SET
			draft      = EXCLUDED.draft,
			updated_at = now()`,
		userID, courseID, levelID, taskID, draft)
	return err
}

// ── Пользователи и сессии ───────────────────────────────────────────────────

// CreateUser заводит пользователя в одной транзакции:
// LOCK гарантирует, что «первый» определяется без гонок; первый = админ,
// и ему передаются прогресс-сироты из БД доавторизационной версии.
func (p *Postgres) CreateUser(username, passwordHash string) (*User, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`LOCK TABLE users IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return nil, err
	}
	var count int
	if err := tx.QueryRow(`SELECT count(*) FROM users`).Scan(&count); err != nil {
		return nil, err
	}
	role := "user"
	if count == 0 {
		role = "admin"
	}

	u := &User{Username: username, PasswordHash: passwordHash, Role: role}
	err = tx.QueryRow(`
		INSERT INTO users (username, password_hash, role)
		VALUES ($1, $2, $3) RETURNING id, created_at`,
		username, passwordHash, role).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, ErrUsernameTaken
		}
		return nil, err
	}

	if count == 0 {
		if _, err := tx.Exec(`UPDATE level_progress SET user_id = $1 WHERE user_id IS NULL`, u.ID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(`UPDATE task_progress SET user_id = $1 WHERE user_id IS NULL`, u.ID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return u, nil
}

func (p *Postgres) UserByName(username string) (*User, error) {
	u := &User{}
	err := p.db.QueryRow(`
		SELECT id, username, password_hash, role, created_at
		  FROM users WHERE username = $1`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (p *Postgres) CreateSession(token string, userID int64, expiresAt time.Time) error {
	_, err := p.db.Exec(`INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, expiresAt)
	return err
}

func (p *Postgres) UserBySession(token string) (*User, time.Time, error) {
	u := &User{}
	var expires time.Time
	err := p.db.QueryRow(`
		SELECT u.id, u.username, u.password_hash, u.role, u.created_at, s.expires_at
		  FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token = $1 AND s.expires_at > now()`, token).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, time.Time{}, nil
	}
	if err != nil {
		return nil, time.Time{}, err
	}
	return u, expires, nil
}

func (p *Postgres) TouchSession(token string, expiresAt time.Time) error {
	_, err := p.db.Exec(`UPDATE sessions SET expires_at = $2 WHERE token = $1`, token, expiresAt)
	return err
}

func (p *Postgres) DeleteSession(token string) error {
	_, err := p.db.Exec(`DELETE FROM sessions WHERE token = $1`, token)
	return err
}

func (p *Postgres) DeleteExpiredSessions() error {
	_, err := p.db.Exec(`DELETE FROM sessions WHERE expires_at <= now()`)
	return err
}

// ── Чат ─────────────────────────────────────────────────────────────────────

const chatKeepRows = 200

func (p *Postgres) ChatHistory(userID int64, courseID, levelID string, limit int) ([]ChatMessage, error) {
	rows, err := p.db.Query(`
		SELECT role, content, created_at FROM (
			SELECT id, role, content, created_at
			  FROM chat_messages
			 WHERE user_id = $1 AND course_id = $2 AND level_id = $3
			 ORDER BY id DESC LIMIT $4
		) t ORDER BY id ASC`, userID, courseID, levelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (p *Postgres) AppendChat(userID int64, courseID, levelID, role, content string) error {
	if _, err := p.db.Exec(`
		INSERT INTO chat_messages (user_id, course_id, level_id, role, content)
		VALUES ($1, $2, $3, $4, $5)`, userID, courseID, levelID, role, content); err != nil {
		return err
	}
	// Прунинг: держим не больше chatKeepRows строк на тему.
	_, err := p.db.Exec(`
		DELETE FROM chat_messages
		 WHERE user_id = $1 AND course_id = $2 AND level_id = $3
		   AND id NOT IN (
			SELECT id FROM chat_messages
			 WHERE user_id = $1 AND course_id = $2 AND level_id = $3
			 ORDER BY id DESC LIMIT $4)`,
		userID, courseID, levelID, chatKeepRows)
	return err
}

func (p *Postgres) ClearChat(userID int64, courseID, levelID string) error {
	_, err := p.db.Exec(`
		DELETE FROM chat_messages
		 WHERE user_id = $1 AND course_id = $2 AND level_id = $3`, userID, courseID, levelID)
	return err
}

// ── Ачивки ──────────────────────────────────────────────────────────────────

func (p *Postgres) Achievements(userID int64) (map[string]time.Time, error) {
	rows, err := p.db.Query(`
		SELECT achievement_id, granted_at FROM user_achievements WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[string]time.Time{}
	for rows.Next() {
		var id string
		var at time.Time
		if err := rows.Scan(&id, &at); err != nil {
			return nil, err
		}
		res[id] = at
	}
	return res, rows.Err()
}

func (p *Postgres) GrantAchievement(userID int64, achievementID string) (bool, error) {
	res, err := p.db.Exec(`
		INSERT INTO user_achievements (user_id, achievement_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, achievementID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// ── Заметки ─────────────────────────────────────────────────────────────────

func (p *Postgres) Note(userID int64, courseID, levelID string) (*Note, error) {
	n := &Note{CourseID: courseID, LevelID: levelID}
	err := p.db.QueryRow(`
		SELECT body, updated_at FROM notes
		 WHERE user_id = $1 AND course_id = $2 AND level_id = $3`,
		userID, courseID, levelID).Scan(&n.Body, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (p *Postgres) SaveNote(userID int64, courseID, levelID, body string) error {
	if strings.TrimSpace(body) == "" {
		_, err := p.db.Exec(`
			DELETE FROM notes WHERE user_id = $1 AND course_id = $2 AND level_id = $3`,
			userID, courseID, levelID)
		return err
	}
	_, err := p.db.Exec(`
		INSERT INTO notes (user_id, course_id, level_id, body)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, course_id, level_id) DO UPDATE SET
			body = EXCLUDED.body, updated_at = now()`,
		userID, courseID, levelID, body)
	return err
}

func (p *Postgres) NotesAll(userID int64) ([]Note, error) {
	rows, err := p.db.Query(`
		SELECT course_id, level_id, body, updated_at FROM notes
		 WHERE user_id = $1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.CourseID, &n.LevelID, &n.Body, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (p *Postgres) Close() error { return p.db.Close() }
