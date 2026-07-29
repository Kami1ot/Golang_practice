package store

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// Memory — хранилище в памяти: для тестов и работы без БД (-dsn "").
// Пользователи и прогресс пропадают при выходе.
type Memory struct {
	mu           sync.Mutex
	levels       map[levelKey]LevelFacts
	tasks        map[taskKey]TaskFacts
	users        map[string]*User // по имени
	nextUserID   int64
	sessions     map[string]memSession
	chats        map[levelKey][]ChatMessage // ключ: (user, course, level)
	achievements map[int64]map[string]time.Time
	notes        map[levelKey]Note
}

type levelKey struct {
	user          int64
	course, level string
}

type taskKey struct {
	user                int64
	course, level, task string
}

type memSession struct {
	userID  int64
	expires time.Time
}

func NewMemory() *Memory {
	return &Memory{
		levels:       map[levelKey]LevelFacts{},
		tasks:        map[taskKey]TaskFacts{},
		users:        map[string]*User{},
		sessions:     map[string]memSession{},
		chats:        map[levelKey][]ChatMessage{},
		achievements: map[int64]map[string]time.Time{},
		notes:        map[levelKey]Note{},
	}
}

// ── Прогресс ────────────────────────────────────────────────────────────────

func (m *Memory) CourseFacts(userID int64, courseID string) (*CourseFacts, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	facts := &CourseFacts{
		Levels: map[string]LevelFacts{},
		Tasks:  map[string]map[string]TaskFacts{},
	}
	for k, v := range m.levels {
		if k.user == userID && k.course == courseID {
			facts.Levels[k.level] = v
		}
	}
	for k, v := range m.tasks {
		if k.user == userID && k.course == courseID {
			if facts.Tasks[k.level] == nil {
				facts.Tasks[k.level] = map[string]TaskFacts{}
			}
			facts.Tasks[k.level][k.task] = v
		}
	}
	return facts, nil
}

func (m *Memory) CourseFactsAll(courseID string) (map[int64]*CourseFacts, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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
	for k, v := range m.levels {
		if k.course == courseID {
			userFacts(k.user).Levels[k.level] = v
		}
	}
	for k, v := range m.tasks {
		if k.course == courseID {
			facts := userFacts(k.user)
			if facts.Tasks[k.level] == nil {
				facts.Tasks[k.level] = map[string]TaskFacts{}
			}
			facts.Tasks[k.level][k.task] = v
		}
	}
	return byUser, nil
}

func (m *Memory) RecordQuizAttempt(userID int64, courseID, levelID string, passed bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := levelKey{userID, courseID, levelID}
	f := m.levels[key]
	f.QuizAttempts++
	f.QuizPassed = f.QuizPassed || passed
	m.levels[key] = f
	return nil
}

func (m *Memory) RecordTaskRun(userID int64, courseID, levelID, taskID, draft string, passed bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := taskKey{userID, courseID, levelID, taskID}
	f := m.tasks[key]
	f.Attempts++
	f.Draft = draft
	if passed && !f.Completed {
		f.Completed = true
		f.CompletedAt = time.Now()
	}
	m.tasks[key] = f
	return nil
}

func (m *Memory) SaveDraft(userID int64, courseID, levelID, taskID, draft string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := taskKey{userID, courseID, levelID, taskID}
	f := m.tasks[key]
	f.Draft = draft
	m.tasks[key] = f
	return nil
}

// ── Пользователи и сессии ───────────────────────────────────────────────────

func (m *Memory) CreateUser(username, passwordHash string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[username]; ok {
		return nil, ErrUsernameTaken
	}
	m.nextUserID++
	role := "user"
	if len(m.users) == 0 {
		role = "admin"
	}
	u := &User{
		ID:           m.nextUserID,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
	}
	m.users[username] = u
	return u, nil
}

func (m *Memory) UserByName(username string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.users[username], nil
}

func (m *Memory) CreateSession(token string, userID int64, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[token] = memSession{userID, expiresAt}
	return nil
}

func (m *Memory) UserBySession(token string) (*User, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[token]
	if !ok || s.expires.Before(time.Now()) {
		return nil, time.Time{}, nil
	}
	for _, u := range m.users {
		if u.ID == s.userID {
			return u, s.expires, nil
		}
	}
	return nil, time.Time{}, nil
}

func (m *Memory) TouchSession(token string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[token]; ok {
		s.expires = expiresAt
		m.sessions[token] = s
	}
	return nil
}

func (m *Memory) DeleteSession(token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
	return nil
}

func (m *Memory) DeleteExpiredSessions() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for t, s := range m.sessions {
		if s.expires.Before(now) {
			delete(m.sessions, t)
		}
	}
	return nil
}

// ── Чат ─────────────────────────────────────────────────────────────────────

func (m *Memory) ChatHistory(userID int64, courseID, levelID string, limit int) ([]ChatMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	all := m.chats[levelKey{userID, courseID, levelID}]
	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	out := make([]ChatMessage, len(all))
	copy(out, all)
	return out, nil
}

func (m *Memory) AppendChat(userID int64, courseID, levelID, role, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := levelKey{userID, courseID, levelID}
	msgs := append(m.chats[key], ChatMessage{Role: role, Content: content, CreatedAt: time.Now()})
	if len(msgs) > chatKeepRows {
		msgs = msgs[len(msgs)-chatKeepRows:]
	}
	m.chats[key] = msgs
	return nil
}

func (m *Memory) ClearChat(userID int64, courseID, levelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.chats, levelKey{userID, courseID, levelID})
	return nil
}

// ── Ачивки ──────────────────────────────────────────────────────────────────

func (m *Memory) Achievements(userID int64) (map[string]time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := map[string]time.Time{}
	for id, at := range m.achievements[userID] {
		res[id] = at
	}
	return res, nil
}

func (m *Memory) GrantAchievement(userID int64, achievementID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.achievements[userID] == nil {
		m.achievements[userID] = map[string]time.Time{}
	}
	if _, ok := m.achievements[userID][achievementID]; ok {
		return false, nil
	}
	m.achievements[userID][achievementID] = time.Now()
	return true, nil
}

// ── Заметки ─────────────────────────────────────────────────────────────────

func (m *Memory) Note(userID int64, courseID, levelID string) (*Note, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.notes[levelKey{userID, courseID, levelID}]
	if !ok {
		return nil, nil
	}
	return &n, nil
}

func (m *Memory) SaveNote(userID int64, courseID, levelID, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := levelKey{userID, courseID, levelID}
	if strings.TrimSpace(body) == "" {
		delete(m.notes, key)
		return nil
	}
	m.notes[key] = Note{CourseID: courseID, LevelID: levelID, Body: body, UpdatedAt: time.Now()}
	return nil
}

func (m *Memory) NotesAll(userID int64) ([]Note, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var notes []Note
	for k, n := range m.notes {
		if k.user == userID {
			notes = append(notes, n)
		}
	}
	sort.Slice(notes, func(i, j int) bool { return notes[i].UpdatedAt.After(notes[j].UpdatedAt) })
	return notes, nil
}

func (m *Memory) Close() error { return nil }
