// Package store хранит прогресс, пользователей, сессии, чат и ачивки.
// Только сырые факты: пройден ли тест, решены ли задачи, черновики, выдачи ачивок.
// Всё производное (замки, доступность, XP, условия ачивок) вычисляется поверх фактов.
package store

import (
	"errors"
	"time"
)

// ErrUsernameTaken возвращается CreateUser при занятом имени.
var ErrUsernameTaken = errors.New("имя пользователя занято")

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string // "admin" | "user"
	CreatedAt    time.Time
}

func (u *User) IsAdmin() bool { return u != nil && u.Role == "admin" }

type ChatMessage struct {
	Role      string // "user" | "assistant"
	Content   string
	CreatedAt time.Time
}

type LevelFacts struct {
	QuizPassed   bool
	QuizAttempts int
}

type TaskFacts struct {
	Completed   bool
	Attempts    int
	Draft       string // "" — черновика нет
	CompletedAt time.Time
}

// Note — заметка пользователя к уровню (Markdown, одна на уровень).
type Note struct {
	CourseID  string
	LevelID   string
	Body      string
	UpdatedAt time.Time
}

// CourseFacts — все факты одного пользователя по одному курсу.
type CourseFacts struct {
	Levels map[string]LevelFacts           // levelID -> факты
	Tasks  map[string]map[string]TaskFacts // levelID -> taskID -> факты
}

func (f *CourseFacts) Level(levelID string) LevelFacts {
	return f.Levels[levelID]
}

func (f *CourseFacts) Task(levelID, taskID string) TaskFacts {
	return f.Tasks[levelID][taskID]
}

type Store interface {
	// Прогресс (все методы — в разрезе пользователя).
	CourseFacts(userID int64, courseID string) (*CourseFacts, error)
	// CourseFactsAll — факты ВСЕХ пользователей по одному курсу (для агрегатов витрины).
	// Ключ — user_id; строки-сироты (user_id IS NULL) не учитываются.
	CourseFactsAll(courseID string) (map[int64]*CourseFacts, error)
	// RecordQuizAttempt инкрементирует счётчик попыток; passed=true фиксируется навсегда.
	RecordQuizAttempt(userID int64, courseID, levelID string, passed bool) error
	// RecordTaskRun сохраняет черновик, инкрементирует попытки; passed=true фиксируется навсегда.
	RecordTaskRun(userID int64, courseID, levelID, taskID, draft string, passed bool) error
	SaveDraft(userID int64, courseID, levelID, taskID, draft string) error

	// Пользователи и сессии.
	// CreateUser: первый пользователь становится админом и усыновляет прогресс-сироты
	// (строки с user_id IS NULL, оставшиеся от версии без авторизации).
	CreateUser(username, passwordHash string) (*User, error)
	UserByName(username string) (*User, error) // (nil, nil), если нет
	CreateSession(token string, userID int64, expiresAt time.Time) error
	UserBySession(token string) (*User, time.Time, error) // (nil, zero, nil), если нет/протухла
	TouchSession(token string, expiresAt time.Time) error  // скользящее продление
	DeleteSession(token string) error
	DeleteExpiredSessions() error

	// Чат с наставником.
	ChatHistory(userID int64, courseID, levelID string, limit int) ([]ChatMessage, error)
	AppendChat(userID int64, courseID, levelID, role, content string) error
	ClearChat(userID int64, courseID, levelID string) error

	// Ачивки: только факты выдачи.
	Achievements(userID int64) (map[string]time.Time, error)
	// GrantAchievement возвращает true, если ачивка выдана только что (а не была раньше).
	GrantAchievement(userID int64, achievementID string) (bool, error)

	// Заметки к уровням.
	Note(userID int64, courseID, levelID string) (*Note, error) // (nil, nil), если нет
	// SaveNote сохраняет заметку; пустое (после TrimSpace) тело удаляет её.
	SaveNote(userID int64, courseID, levelID, body string) error
	NotesAll(userID int64) ([]Note, error) // все заметки пользователя, свежие сверху

	Close() error
}
