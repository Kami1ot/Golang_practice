package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

	"gopractice/internal/content"
	"gopractice/internal/runner"
	"gopractice/internal/store"
)

func (a *API) handleCourses(w http.ResponseWriter, r *http.Request) {
	cat := a.cat()
	u := userFrom(r)
	courses := []courseDTO{}
	for _, id := range cat.Order {
		c := cat.Courses[id]
		facts, err := a.store.CourseFacts(u.ID, id)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
			return
		}
		completed, xpEarned, xpTotal := courseStats(c, facts)
		agg, err := a.courseAggregates(c)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
			return
		}
		tags := c.Tags
		if tags == nil {
			tags = []string{}
		}
		courses = append(courses, courseDTO{
			ID:              c.ID,
			Title:           c.Title,
			Description:     c.Description,
			Tags:            tags,
			LevelsTotal:     len(c.Levels),
			LevelsCompleted: completed,
			XPEarned:        xpEarned,
			XPTotal:         xpTotal,
			UsersCompleted:  agg.Completed,
			UsersInProgress: agg.InProgress,
		})
	}
	writeJSON(w, http.StatusOK, courses)
}

// courseFromPath достаёт курс из {course} или пишет 404.
func (a *API) courseFromPath(w http.ResponseWriter, r *http.Request) (*content.Course, *store.CourseFacts, bool) {
	c, ok := a.cat().Courses[r.PathValue("course")]
	if !ok {
		writeErr(w, http.StatusNotFound, "курс не найден")
		return nil, nil, false
	}
	facts, err := a.store.CourseFacts(userFrom(r).ID, c.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return nil, nil, false
	}
	return c, facts, true
}

// levelFromPath достаёт уровень из {level} с проверкой замка.
func (a *API) levelFromPath(w http.ResponseWriter, r *http.Request) (*content.Course, *content.Level, *store.CourseFacts, bool) {
	c, facts, ok := a.courseFromPath(w, r)
	if !ok {
		return nil, nil, nil, false
	}
	l, ok := c.Levels[r.PathValue("level")]
	if !ok {
		writeErr(w, http.StatusNotFound, "уровень не найден")
		return nil, nil, nil, false
	}
	if deriveStates(c, facts)[l.ID] == stateLocked {
		writeErr(w, http.StatusForbidden, "уровень закрыт: сначала пройдите предыдущие")
		return nil, nil, nil, false
	}
	return c, l, facts, true
}

func (a *API) handleTree(w http.ResponseWriter, r *http.Request) {
	c, facts, ok := a.courseFromPath(w, r)
	if !ok {
		return
	}
	states := deriveStates(c, facts)
	_, xpEarned, xpTotal := courseStats(c, facts)

	tree := treeDTO{
		Course:   courseRefDTO{ID: c.ID, Title: c.Title},
		XPEarned: xpEarned,
		XPTotal:  xpTotal,
		Nodes:    []treeNodeDTO{},
	}
	for _, id := range c.LevelOrder {
		l := c.Levels[id]
		tasksDone := 0
		required := l.RequiredTaskIDs()
		for _, tid := range required {
			if facts.Task(id, tid).Completed {
				tasksDone++
			}
		}
		requires := l.Requires
		if requires == nil {
			requires = []string{}
		}
		tree.Nodes = append(tree.Nodes, treeNodeDTO{
			ID:       l.ID,
			Title:    l.Title,
			Summary:  l.Summary,
			Requires: requires,
			Optional: l.Optional,
			XP:       l.XP,
			Group:    l.Group,
			State:    states[id],
			Progress: nodeProgress{
				HasQuiz:       l.Quiz != nil,
				QuizPassed:    facts.Level(id).QuizPassed,
				TasksDone:     tasksDone,
				TasksRequired: len(required),
			},
		})
	}
	writeJSON(w, http.StatusOK, tree)
}

func (a *API) handleLevel(w http.ResponseWriter, r *http.Request) {
	c, l, facts, ok := a.levelFromPath(w, r)
	if !ok {
		return
	}
	dto := levelDTO{
		ID:         l.ID,
		Title:      l.Title,
		Summary:    l.Summary,
		Optional:   l.Optional,
		XP:         l.XP,
		State:      deriveStates(c, facts)[l.ID],
		LessonMD:   l.LessonMD,
		Quiz:       toQuizDTO(l.Quiz),
		QuizPassed: facts.Level(l.ID).QuizPassed,
		Tasks:      []taskDTO{},
	}
	for _, t := range l.Tasks {
		dto.Tasks = append(dto.Tasks, toTaskDTO(t, facts.Task(l.ID, t.ID)))
	}
	writeJSON(w, http.StatusOK, dto)
}

type quizSubmission struct {
	Answers map[string]json.RawMessage `json:"answers"`
}

type quizResultDTO struct {
	Passed          bool                         `json:"passed"`
	Score           float64                      `json:"score"`
	LevelCompleted  bool                         `json:"levelCompleted"`
	Results         map[string]questionResultDTO `json:"results"`
	NewAchievements []achievementDTO             `json:"newAchievements,omitempty"`
}

type questionResultDTO struct {
	Correct     bool   `json:"correct"`
	Explanation string `json:"explanation,omitempty"`
}

func (a *API) handleQuiz(w http.ResponseWriter, r *http.Request) {
	c, l, _, ok := a.levelFromPath(w, r)
	if !ok {
		return
	}
	if l.Quiz == nil {
		writeErr(w, http.StatusNotFound, "у уровня нет теста")
		return
	}
	var sub quizSubmission
	if !readJSON(w, r, &sub) {
		return
	}

	results := map[string]questionResultDTO{}
	correct := 0
	for _, q := range l.Quiz.Questions {
		ok := gradeQuestion(q, sub.Answers[q.ID])
		if ok {
			correct++
		}
		results[q.ID] = questionResultDTO{Correct: ok, Explanation: q.Explanation}
	}
	score := float64(correct) / float64(len(l.Quiz.Questions))
	passed := score >= l.Quiz.PassScore-1e-9

	u := userFrom(r)
	if err := a.store.RecordQuizAttempt(u.ID, c.ID, l.ID, passed); err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	facts, err := a.store.CourseFacts(u.ID, c.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, quizResultDTO{
		Passed:         passed,
		Score:          score,
		LevelCompleted: completedSet(c, facts)[l.ID],
		Results:        results,
		NewAchievements: a.evaluateAchievements(u),
	})
}

func gradeQuestion(q *content.Question, raw json.RawMessage) bool {
	if raw == nil {
		return false
	}
	switch q.Type {
	case content.QSingle:
		var idx int
		return json.Unmarshal(raw, &idx) == nil && idx == q.AnswerIndex
	case content.QMulti:
		var idxs []int
		if json.Unmarshal(raw, &idxs) != nil {
			return false
		}
		sort.Ints(idxs)
		if len(idxs) != len(q.AnswerIndices) {
			return false
		}
		for i := range idxs {
			if idxs[i] != q.AnswerIndices[i] {
				return false
			}
		}
		return true
	case content.QOutput:
		var s string
		return json.Unmarshal(raw, &s) == nil &&
			runner.Normalize(s) == runner.Normalize(q.AnswerText)
	case content.QBlank:
		var s string
		if json.Unmarshal(raw, &s) != nil {
			return false
		}
		s = strings.TrimSpace(s)
		for _, accepted := range q.AnswerTexts {
			accepted = strings.TrimSpace(accepted)
			if q.IsCaseSensitive() {
				if s == accepted {
					return true
				}
			} else if strings.EqualFold(s, accepted) {
				return true
			}
		}
	}
	return false
}

type runSubmission struct {
	Code string `json:"code"`
}

type runResultDTO struct {
	Status          string           `json:"status"`
	CompileOutput   string           `json:"compileOutput,omitempty"`
	Tests           []testResultDTO  `json:"tests"`
	TaskCompleted   bool             `json:"taskCompleted"`
	LevelCompleted  bool             `json:"levelCompleted"`
	NewAchievements []achievementDTO `json:"newAchievements,omitempty"`
}

type testResultDTO struct {
	Name       string `json:"name"`
	Passed     bool   `json:"passed"`
	Hidden     bool   `json:"hidden"`
	Skipped    bool   `json:"skipped,omitempty"`
	Stdin      string `json:"stdin,omitempty"`
	Expected   string `json:"expected,omitempty"`
	Actual     string `json:"actual,omitempty"`
	Output     string `json:"output,omitempty"`
	DurationMs int64  `json:"durationMs"`
}

func (a *API) handleRun(w http.ResponseWriter, r *http.Request) {
	c, l, _, ok := a.levelFromPath(w, r)
	if !ok {
		return
	}
	t := l.Task(r.PathValue("task"))
	if t == nil {
		writeErr(w, http.StatusNotFound, "задача не найдена")
		return
	}
	var sub runSubmission
	if !readJSON(w, r, &sub) {
		return
	}
	if strings.TrimSpace(sub.Code) == "" {
		writeErr(w, http.StatusBadRequest, "пустой код")
		return
	}

	res, err := a.runner.Run(r.Context(), t, sub.Code)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "запуск: "+err.Error())
		return
	}
	passed := res.Status == runner.StatusPassed

	u := userFrom(r)
	if err := a.store.RecordTaskRun(u.ID, c.ID, l.ID, t.ID, sub.Code, passed); err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	facts, err := a.store.CourseFacts(u.ID, c.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}

	dto := runResultDTO{
		Status:         res.Status,
		CompileOutput:  res.CompileOutput,
		Tests:          []testResultDTO{},
		TaskCompleted:  facts.Task(l.ID, t.ID).Completed,
		LevelCompleted: completedSet(c, facts)[l.ID],
	}
	for _, tr := range res.Tests {
		item := testResultDTO{
			Name:       tr.Name,
			Passed:     tr.Passed,
			Hidden:     tr.Hidden,
			Skipped:    tr.Skipped,
			DurationMs: tr.DurationMs,
		}
		// Скрытые тесты не раскрывают данные — только вердикт.
		if !tr.Hidden {
			item.Stdin = tr.Stdin
			item.Expected = tr.Expected
			item.Actual = tr.Actual
			item.Output = tr.Output
		}
		dto.Tests = append(dto.Tests, item)
	}
	dto.NewAchievements = a.evaluateAchievements(u)
	writeJSON(w, http.StatusOK, dto)
}

type draftSubmission struct {
	Code string `json:"code"`
}

func (a *API) handleDraft(w http.ResponseWriter, r *http.Request) {
	c, l, _, ok := a.levelFromPath(w, r)
	if !ok {
		return
	}
	t := l.Task(r.PathValue("task"))
	if t == nil {
		writeErr(w, http.StatusNotFound, "задача не найдена")
		return
	}
	var sub draftSubmission
	if !readJSON(w, r, &sub) {
		return
	}
	if err := a.store.SaveDraft(userFrom(r).ID, c.ID, l.ID, t.ID, sub.Code); err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// reloadCatalog перечитывает контент с диска; при ошибках валидации старый
// каталог продолжает работать, а вызывающему возвращается список проблем.
// Вызывать ТОЛЬКО под a.fsMu.
func (a *API) reloadCatalog() []string {
	catalog, err := content.Load(a.contentDir)
	if err != nil {
		return strings.Split(err.Error(), "\n")
	}
	a.mu.Lock()
	a.catalog = catalog
	a.mu.Unlock()
	a.invalidateAggregates() // состав уровней мог измениться — «завершённость» тоже
	log.Printf("контент перезагружен: курсов — %d", len(catalog.Order))
	return nil
}

func (a *API) handleReload(w http.ResponseWriter, r *http.Request) {
	a.fsMu.Lock()
	errs := a.reloadCatalog()
	a.fsMu.Unlock()
	if errs != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"ok": false, "errors": errs})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
