package httpapi

import (
	"gopractice/internal/content"
	"gopractice/internal/store"
)

// Состояния уровня — всегда вычисляются из фактов, никогда не хранятся.
const (
	stateLocked    = "locked"
	stateAvailable = "available"
	stateCompleted = "completed"
)

// ownWorkDone: тест пройден (или его нет) И все обязательные задачи решены.
// Только собственная работа уровня, без учёта предпосылок.
func ownWorkDone(l *content.Level, facts *store.CourseFacts) bool {
	if l.Quiz != nil && !facts.Level(l.ID).QuizPassed {
		return false
	}
	for _, id := range l.RequiredTaskIDs() {
		if !facts.Task(l.ID, id).Completed {
			return false
		}
	}
	return true
}

// completedSet: уровень завершён, когда завершены все его предпосылки
// И выполнена собственная работа. Транзитивность важна для уровней без
// теста и задач — иначе они «завершались» бы ещё до разблокировки.
func completedSet(c *content.Course, facts *store.CourseFacts) map[string]bool {
	done := make(map[string]bool, len(c.Levels))
	var isDone func(id string) bool
	isDone = func(id string) bool {
		if v, ok := done[id]; ok {
			return v
		}
		l := c.Levels[id]
		result := ownWorkDone(l, facts)
		for _, req := range l.Requires {
			if !result {
				break
			}
			result = isDone(req)
		}
		done[id] = result // цикличность исключена валидатором контента
		return result
	}
	for id := range c.Levels {
		isDone(id)
	}
	return done
}

// courseCompletedFor: курс завершён, когда завершены все обязательные уровни
// (та же семантика, что у ачивки «Курс завершён»; бонусные уровни не требуются).
func courseCompletedFor(c *content.Course, facts *store.CourseFacts) bool {
	if len(c.Levels) == 0 {
		return false
	}
	done := completedSet(c, facts)
	for _, l := range c.Levels {
		if !l.Optional && !done[l.ID] {
			return false
		}
	}
	return true
}

// deriveStates вычисляет состояние каждого уровня курса.
func deriveStates(c *content.Course, facts *store.CourseFacts) map[string]string {
	completed := completedSet(c, facts)
	states := make(map[string]string, len(c.Levels))
	for id, l := range c.Levels {
		switch {
		case completed[id]:
			states[id] = stateCompleted
		default:
			states[id] = stateAvailable
			for _, req := range l.Requires {
				if !completed[req] {
					states[id] = stateLocked
					break
				}
			}
		}
	}
	return states
}

type courseDTO struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Tags            []string `json:"tags"`
	LevelsTotal     int      `json:"levelsTotal"`
	LevelsCompleted int      `json:"levelsCompleted"`
	XPEarned        int      `json:"xpEarned"`
	XPTotal         int      `json:"xpTotal"`
	UsersCompleted  int      `json:"usersCompleted"`
	UsersInProgress int      `json:"usersInProgress"`
}

type treeDTO struct {
	Course   courseRefDTO  `json:"course"`
	XPEarned int           `json:"xpEarned"`
	XPTotal  int           `json:"xpTotal"`
	Nodes    []treeNodeDTO `json:"nodes"`
}

type courseRefDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type treeNodeDTO struct {
	ID       string       `json:"id"`
	Title    string       `json:"title"`
	Summary  string       `json:"summary"`
	Requires []string     `json:"requires"`
	Optional bool         `json:"optional"`
	XP       int          `json:"xp"`
	Group    string       `json:"group,omitempty"`
	State    string       `json:"state"`
	Progress nodeProgress `json:"progress"`
}

type nodeProgress struct {
	HasQuiz       bool `json:"hasQuiz"`
	QuizPassed    bool `json:"quizPassed"`
	TasksDone     int  `json:"tasksDone"`
	TasksRequired int  `json:"tasksRequired"`
}

type levelDTO struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary"`
	Optional   bool      `json:"optional"`
	XP         int       `json:"xp"`
	State      string    `json:"state"`
	LessonMD   string    `json:"lessonMd"`
	Quiz       *quizDTO  `json:"quiz,omitempty"`
	QuizPassed bool      `json:"quizPassed"`
	Tasks      []taskDTO `json:"tasks"`
}

// quizDTO — вопросы БЕЗ ответов и объяснений: они уезжают клиенту
// только после проверки на сервере.
type quizDTO struct {
	PassScore float64       `json:"passScore"`
	Questions []questionDTO `json:"questions"`
}

type questionDTO struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options,omitempty"`
	Code    string   `json:"code,omitempty"`
}

type taskDTO struct {
	ID          string       `json:"id"`
	Type        string       `json:"type"`
	Title       string       `json:"title"`
	Required    bool         `json:"required"`
	TimeoutSec  int          `json:"timeoutSec"`
	StatementMD string       `json:"statementMd"`
	Code        string       `json:"code"`        // черновик, если есть, иначе заготовка
	StarterCode string       `json:"starterCode"` // для «Сбросить к заготовке»
	Completed   bool         `json:"completed"`
	Attempts    int          `json:"attempts"`
	Examples    []exampleDTO `json:"examples,omitempty"` // только видимые тест-кейсы
	Hints       []string     `json:"hints,omitempty"`    // прогрессивные подсказки
}

type exampleDTO struct {
	Name   string `json:"name"`
	Stdin  string `json:"stdin"`
	Stdout string `json:"stdout"`
}

func toQuizDTO(q *content.Quiz) *quizDTO {
	if q == nil {
		return nil
	}
	dto := &quizDTO{PassScore: q.PassScore}
	for _, question := range q.Questions {
		dto.Questions = append(dto.Questions, questionDTO{
			ID:      question.ID,
			Type:    question.Type,
			Prompt:  question.Prompt,
			Options: question.Options,
			Code:    question.Code,
		})
	}
	return dto
}

func toTaskDTO(t *content.Task, facts store.TaskFacts) taskDTO {
	code := facts.Draft
	if code == "" {
		code = t.StarterCode
	}
	dto := taskDTO{
		ID:          t.ID,
		Type:        t.Type,
		Title:       t.Title,
		Required:    t.IsRequired(),
		TimeoutSec:  t.Timeout(),
		StatementMD: t.StatementMD,
		Code:        code,
		StarterCode: t.StarterCode,
		Completed:   facts.Completed,
		Attempts:    facts.Attempts,
		Hints:       t.Hints,
	}
	for _, tc := range t.Tests {
		if !tc.Hidden {
			dto.Examples = append(dto.Examples, exampleDTO{Name: tc.Name, Stdin: tc.Stdin, Stdout: tc.Stdout})
		}
	}
	return dto
}

func courseStats(c *content.Course, facts *store.CourseFacts) (completed, xpEarned, xpTotal int) {
	done := completedSet(c, facts)
	for _, l := range c.Levels {
		xpTotal += l.XP
		if done[l.ID] {
			completed++
			xpEarned += l.XP
		}
	}
	return completed, xpEarned, xpTotal
}
