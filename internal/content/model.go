package content

import "encoding/json"

// Catalog — все загруженные курсы. Пересоздаётся целиком при reload.
type Catalog struct {
	Courses map[string]*Course
	Order   []string // порядок курсов по имени папки
}

type Course struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"` // ключевые слова для витрины курсов
	LevelOrder  []string `json:"levels"`

	Levels map[string]*Level `json:"-"`
	Dir    string            `json:"-"`
}

type Level struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	Requires []string `json:"requires"`
	Optional bool     `json:"optional"`
	XP       int      `json:"xp"`
	Group    string   `json:"group,omitempty"` // секция на карте уровней («Основы», «Циклы»…)

	LessonMD string  `json:"-"`
	Quiz     *Quiz   `json:"-"`
	Tasks    []*Task `json:"-"`
	Dir      string  `json:"-"`
}

type Quiz struct {
	PassScore float64     `json:"passScore"`
	Questions []*Question `json:"questions"`
}

// Типы вопросов.
const (
	QSingle = "single" // один вариант, answer — индекс
	QMulti  = "multi"  // несколько вариантов, answers — точное множество индексов
	QOutput = "output" // «что выведет код», answer — строка, сравнение через normalize
	QBlank  = "blank"  // пропуск, answers — допустимые строки
)

type Question struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Prompt        string          `json:"prompt"`
	Options       []string        `json:"options,omitempty"`
	Code          string          `json:"code,omitempty"`
	CodeFile      string          `json:"codeFile,omitempty"`
	RawAnswer     json.RawMessage `json:"answer,omitempty"`
	RawAnswers    json.RawMessage `json:"answers,omitempty"`
	CaseSensitive *bool           `json:"caseSensitive,omitempty"`
	Explanation   string          `json:"explanation,omitempty"`

	// Типизированные ответы, заполняются загрузчиком в зависимости от Type.
	AnswerIndex   int      `json:"-"` // single
	AnswerIndices []int    `json:"-"` // multi (отсортированы)
	AnswerText    string   `json:"-"` // output
	AnswerTexts   []string `json:"-"` // blank
}

// IsCaseSensitive — для blank-вопросов; по умолчанию true.
func (q *Question) IsCaseSensitive() bool {
	return q.CaseSensitive == nil || *q.CaseSensitive
}

// Типы задач.
const (
	TaskIO       = "io"       // программа целиком, сравнение stdout по тест-кейсам
	TaskUnitTest = "unittest" // функции + скрытый main_test.go, go test
)

type TestCase struct {
	Name   string `json:"name"`
	Stdin  string `json:"stdin"`
	Stdout string `json:"stdout"`
	Hidden bool   `json:"hidden"`
}

type Task struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Required   *bool      `json:"required,omitempty"` // по умолчанию true
	Order      int        `json:"order"`
	TimeoutSec int        `json:"timeoutSec,omitempty"` // по умолчанию 10
	Tests      []TestCase `json:"tests,omitempty"`
	Hints      []string   `json:"hints,omitempty"` // прогрессивные подсказки, от лёгкой к почти-решению

	StatementMD string `json:"-"`
	StarterCode string `json:"-"`
	TestFile    string `json:"-"` // unittest: содержимое main_test.go, клиенту не отдаётся
	Dir         string `json:"-"`
}

func (t *Task) IsRequired() bool {
	return t.Required == nil || *t.Required
}

func (t *Task) Timeout() int {
	if t.TimeoutSec <= 0 {
		return 10
	}
	return t.TimeoutSec
}

// RequiredTaskIDs — id обязательных задач уровня.
func (l *Level) RequiredTaskIDs() []string {
	var ids []string
	for _, t := range l.Tasks {
		if t.IsRequired() {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

// Task возвращает задачу уровня по id или nil.
func (l *Level) Task(id string) *Task {
	for _, t := range l.Tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}
