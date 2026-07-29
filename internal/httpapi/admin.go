package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopractice/internal/content"
)

// Админ-редактор контента. Контент остаётся файлами на диске — единственным
// источником правды; каждая мутация = запись файлов → валидация существующим
// content.Load → hot-swap каталога, при ошибках — побайтовый откат.

var (
	adminIDRe    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	imageNameRe  = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
	winReserved  = map[string]bool{
		"con": true, "prn": true, "aux": true, "nul": true,
		"com1": true, "com2": true, "com3": true, "com4": true, "com5": true,
		"com6": true, "com7": true, "com8": true, "com9": true,
		"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true, "lpt5": true,
		"lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
	}
)

func validID(id string) bool {
	return adminIDRe.MatchString(id) && !winReserved[id]
}

// pathID достаёт и валидирует id из пути. Пустая строка = ошибка уже записана.
func pathID(w http.ResponseWriter, r *http.Request, key string) string {
	id := r.PathValue(key)
	if !validID(id) {
		writeErr(w, http.StatusBadRequest, "некорректный id "+key)
		return ""
	}
	return id
}

func (a *API) registerAdminRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/admin/courses", a.requireAdmin(a.adminCourseList))
	mux.Handle("POST /api/admin/courses", a.requireAdmin(a.adminCourseCreate))
	mux.Handle("GET /api/admin/courses/{course}", a.requireAdmin(a.adminCourseGet))
	mux.Handle("PUT /api/admin/courses/{course}", a.requireAdmin(a.adminCourseUpdate))
	mux.Handle("DELETE /api/admin/courses/{course}", a.requireAdmin(a.adminCourseDelete))

	mux.Handle("POST /api/admin/courses/{course}/levels", a.requireAdmin(a.adminLevelCreate))
	mux.Handle("GET /api/admin/courses/{course}/levels/{level}", a.requireAdmin(a.adminLevelGet))
	mux.Handle("PUT /api/admin/courses/{course}/levels/{level}", a.requireAdmin(a.adminLevelUpdate))
	mux.Handle("DELETE /api/admin/courses/{course}/levels/{level}", a.requireAdmin(a.adminLevelDelete))

	mux.Handle("PUT /api/admin/courses/{course}/levels/{level}/lesson", a.requireAdmin(a.adminLessonPut))
	mux.Handle("PUT /api/admin/courses/{course}/levels/{level}/quiz", a.requireAdmin(a.adminQuizPut))
	mux.Handle("DELETE /api/admin/courses/{course}/levels/{level}/quiz", a.requireAdmin(a.adminQuizDelete))

	mux.Handle("POST /api/admin/courses/{course}/levels/{level}/tasks", a.requireAdmin(a.adminTaskUpsert))
	mux.Handle("PUT /api/admin/courses/{course}/levels/{level}/tasks/{task}", a.requireAdmin(a.adminTaskUpsert))
	mux.Handle("DELETE /api/admin/courses/{course}/levels/{level}/tasks/{task}", a.requireAdmin(a.adminTaskDelete))

	mux.Handle("POST /api/admin/courses/{course}/levels/{level}/images", a.requireAdmin(a.adminImageUpload))
	mux.Handle("DELETE /api/admin/courses/{course}/levels/{level}/images/{name}", a.requireAdmin(a.adminImageDelete))
}

// ── Файловые примитивы ──────────────────────────────────────────────────────

func (a *API) coursePath(course string) string {
	return filepath.Join(a.contentDir, "courses", course)
}

func (a *API) levelPath(course, level string) string {
	return filepath.Join(a.coursePath(course), "levels", level)
}

func writeJSONFile(snap *fsSnapshot, path string, v any) error {
	if err := snap.rememberFile(path); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func writeTextFile(snap *fsSnapshot, path, text string) error {
	if err := snap.rememberFile(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

// httpError — ошибка мутации с HTTP-статусом.
type httpError struct {
	status int
	msg    string
}

func (e *httpError) Error() string { return e.msg }

func httpErrf(status int, format string, args ...any) error {
	return &httpError{status, fmt.Sprintf(format, args...)}
}

// mutateContent — общий каркас мутации: fsMu → снапшот → запись → валидация → откат/фиксация.
func (a *API) mutateContent(w http.ResponseWriter, apply func(snap *fsSnapshot) error) {
	a.fsMu.Lock()
	defer a.fsMu.Unlock()

	snap := newSnapshot(a.contentDir)
	if err := apply(snap); err != nil {
		snap.restore()
		var he *httpError
		if errors.As(err, &he) {
			writeErr(w, he.status, he.msg)
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if errs := a.reloadCatalog(); errs != nil {
		snap.restore()
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"ok": false, "errors": errs})
		return
	}
	snap.discard()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Дисковые формы JSON-файлов (порядок полей стабилен для читаемых диффов).
type courseJSON struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Levels      []string `json:"levels"`
}

type levelJSON struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	Requires []string `json:"requires"`
	Optional bool     `json:"optional,omitempty"`
	XP       int      `json:"xp"`
	Group    string   `json:"group,omitempty"`
}

type taskJSON struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Title      string             `json:"title"`
	Required   bool               `json:"required"`
	Order      int                `json:"order"`
	TimeoutSec int                `json:"timeoutSec,omitempty"`
	Tests      []content.TestCase `json:"tests,omitempty"`
	Hints      []string           `json:"hints,omitempty"`
}

func readJSONFile[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	v := new(T)
	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}
	return v, nil
}

// ── Курсы ───────────────────────────────────────────────────────────────────

func (a *API) adminCourseList(w http.ResponseWriter, r *http.Request) {
	coursesDir := filepath.Join(a.contentDir, "courses")
	entries, _ := os.ReadDir(coursesDir)
	list := []courseJSON{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		c, err := readJSONFile[courseJSON](filepath.Join(coursesDir, e.Name(), "course.json"))
		if err != nil {
			continue
		}
		list = append(list, *c)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	writeJSON(w, http.StatusOK, list)
}

func (a *API) adminCourseCreate(w http.ResponseWriter, r *http.Request) {
	var req courseJSON
	if !readJSON(w, r, &req) {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		if !validID(req.ID) {
			return httpErrf(http.StatusBadRequest, "некорректный id курса (строчные латинские, цифры, дефис)")
		}
		dir := a.coursePath(req.ID)
		if _, err := os.Stat(dir); err == nil {
			return httpErrf(http.StatusConflict, "курс %q уже существует", req.ID)
		}
		req.Levels = []string{}
		return writeJSONFile(snap, filepath.Join(dir, "course.json"), req)
	})
}

func (a *API) adminCourseGet(w http.ResponseWriter, r *http.Request) {
	course := pathID(w, r, "course")
	if course == "" {
		return
	}
	c, err := readJSONFile[courseJSON](filepath.Join(a.coursePath(course), "course.json"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "курс не найден")
		return
	}
	type levelBrief struct {
		levelJSON
		HasQuiz bool `json:"hasQuiz"`
		Tasks   int  `json:"tasks"`
	}
	levels := []levelBrief{}
	for _, id := range c.Levels {
		lv, err := readJSONFile[levelJSON](filepath.Join(a.levelPath(course, id), "level.json"))
		if err != nil {
			continue
		}
		brief := levelBrief{levelJSON: *lv}
		if _, err := os.Stat(filepath.Join(a.levelPath(course, id), "quiz.json")); err == nil {
			brief.HasQuiz = true
		}
		if entries, err := os.ReadDir(filepath.Join(a.levelPath(course, id), "tasks")); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					brief.Tasks++
				}
			}
		}
		levels = append(levels, brief)
	}
	writeJSON(w, http.StatusOK, map[string]any{"course": c, "levels": levels})
}

func (a *API) adminCourseUpdate(w http.ResponseWriter, r *http.Request) {
	course := pathID(w, r, "course")
	if course == "" {
		return
	}
	var req courseJSON
	if !readJSON(w, r, &req) {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		path := filepath.Join(a.coursePath(course), "course.json")
		old, err := readJSONFile[courseJSON](path)
		if err != nil {
			return httpErrf(http.StatusNotFound, "курс не найден")
		}
		old.Title, old.Description = req.Title, req.Description
		if req.Tags != nil {
			old.Tags = req.Tags
		}
		if req.Levels != nil {
			for _, id := range req.Levels {
				if !validID(id) {
					return httpErrf(http.StatusBadRequest, "некорректный id уровня %q", id)
				}
			}
			old.Levels = req.Levels
		}
		return writeJSONFile(snap, path, old)
	})
}

func (a *API) adminCourseDelete(w http.ResponseWriter, r *http.Request) {
	course := pathID(w, r, "course")
	if course == "" {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		dir := a.coursePath(course)
		if _, err := os.Stat(dir); err != nil {
			return httpErrf(http.StatusNotFound, "курс не найден")
		}
		return snap.moveDirToTrash(dir)
	})
}

// ── Уровни ──────────────────────────────────────────────────────────────────

func (a *API) adminLevelCreate(w http.ResponseWriter, r *http.Request) {
	course := pathID(w, r, "course")
	if course == "" {
		return
	}
	var req levelJSON
	if !readJSON(w, r, &req) {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		if !validID(req.ID) {
			return httpErrf(http.StatusBadRequest, "некорректный id уровня")
		}
		coursePath := filepath.Join(a.coursePath(course), "course.json")
		c, err := readJSONFile[courseJSON](coursePath)
		if err != nil {
			return httpErrf(http.StatusNotFound, "курс не найден")
		}
		dir := a.levelPath(course, req.ID)
		if _, err := os.Stat(dir); err == nil {
			return httpErrf(http.StatusConflict, "уровень %q уже существует", req.ID)
		}
		if req.Requires == nil {
			req.Requires = []string{}
		}
		if err := writeJSONFile(snap, filepath.Join(dir, "level.json"), req); err != nil {
			return err
		}
		stub := "# " + req.Title + "\n\nТекст урока пока не написан.\n"
		if err := writeTextFile(snap, filepath.Join(dir, "lesson.md"), stub); err != nil {
			return err
		}
		c.Levels = append(c.Levels, req.ID)
		return writeJSONFile(snap, coursePath, c)
	})
}

type adminTaskDTO struct {
	taskJSON
	StatementMD string `json:"statementMd"`
	StarterCode string `json:"starterCode"`
	TestFile    string `json:"testFile,omitempty"`
}

func (a *API) adminLevelGet(w http.ResponseWriter, r *http.Request) {
	course, level := pathID(w, r, "course"), r.PathValue("level")
	if course == "" {
		return
	}
	if !validID(level) {
		writeErr(w, http.StatusBadRequest, "некорректный id level")
		return
	}
	dir := a.levelPath(course, level)
	meta, err := readJSONFile[levelJSON](filepath.Join(dir, "level.json"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "уровень не найден")
		return
	}
	lesson, _ := os.ReadFile(filepath.Join(dir, "lesson.md"))

	resp := map[string]any{
		"meta":     meta,
		"lessonMd": string(lesson),
	}
	if quiz, err := os.ReadFile(filepath.Join(dir, "quiz.json")); err == nil {
		resp["quiz"] = json.RawMessage(quiz)
	}

	tasks := []adminTaskDTO{}
	if entries, err := os.ReadDir(filepath.Join(dir, "tasks")); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			tdir := filepath.Join(dir, "tasks", e.Name())
			tj, err := readJSONFile[taskJSON](filepath.Join(tdir, "task.json"))
			if err != nil {
				continue
			}
			dto := adminTaskDTO{taskJSON: *tj}
			if b, err := os.ReadFile(filepath.Join(tdir, "statement.md")); err == nil {
				dto.StatementMD = string(b)
			}
			if b, err := os.ReadFile(filepath.Join(tdir, "starter.go")); err == nil {
				dto.StarterCode = string(b)
			}
			if b, err := os.ReadFile(filepath.Join(tdir, "main_test.go")); err == nil {
				dto.TestFile = string(b)
			}
			tasks = append(tasks, dto)
		}
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Order < tasks[j].Order })
	resp["tasks"] = tasks

	images := []string{}
	if entries, err := os.ReadDir(filepath.Join(dir, "img")); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				images = append(images, e.Name())
			}
		}
	}
	resp["images"] = images
	writeJSON(w, http.StatusOK, resp)
}

func (a *API) adminLevelUpdate(w http.ResponseWriter, r *http.Request) {
	course, level := pathID(w, r, "course"), r.PathValue("level")
	if course == "" {
		return
	}
	var req levelJSON
	if !readJSON(w, r, &req) {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		if !validID(level) {
			return httpErrf(http.StatusBadRequest, "некорректный id уровня")
		}
		path := filepath.Join(a.levelPath(course, level), "level.json")
		if _, err := os.Stat(path); err != nil {
			return httpErrf(http.StatusNotFound, "уровень не найден")
		}
		req.ID = level // id менять нельзя
		if req.Requires == nil {
			req.Requires = []string{}
		}
		return writeJSONFile(snap, path, req)
	})
}

func (a *API) adminLevelDelete(w http.ResponseWriter, r *http.Request) {
	course, level := pathID(w, r, "course"), r.PathValue("level")
	if course == "" {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		if !validID(level) {
			return httpErrf(http.StatusBadRequest, "некорректный id уровня")
		}
		coursePath := filepath.Join(a.coursePath(course), "course.json")
		c, err := readJSONFile[courseJSON](coursePath)
		if err != nil {
			return httpErrf(http.StatusNotFound, "курс не найден")
		}
		kept := c.Levels[:0]
		for _, id := range c.Levels {
			if id != level {
				kept = append(kept, id)
			}
		}
		c.Levels = kept
		if err := writeJSONFile(snap, coursePath, c); err != nil {
			return err
		}
		return snap.moveDirToTrash(a.levelPath(course, level))
	})
}

// ── Урок, тест, задачи ──────────────────────────────────────────────────────

func (a *API) adminLessonPut(w http.ResponseWriter, r *http.Request) {
	course, level := pathID(w, r, "course"), r.PathValue("level")
	if course == "" {
		return
	}
	var req struct {
		Markdown string `json:"markdown"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		if !validID(level) {
			return httpErrf(http.StatusBadRequest, "некорректный id уровня")
		}
		dir := a.levelPath(course, level)
		if _, err := os.Stat(dir); err != nil {
			return httpErrf(http.StatusNotFound, "уровень не найден")
		}
		return writeTextFile(snap, filepath.Join(dir, "lesson.md"), req.Markdown)
	})
}

func (a *API) adminQuizPut(w http.ResponseWriter, r *http.Request) {
	course, level := pathID(w, r, "course"), r.PathValue("level")
	if course == "" {
		return
	}
	var raw json.RawMessage
	if !readJSON(w, r, &raw) {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		if !validID(level) {
			return httpErrf(http.StatusBadRequest, "некорректный id уровня")
		}
		dir := a.levelPath(course, level)
		if _, err := os.Stat(dir); err != nil {
			return httpErrf(http.StatusNotFound, "уровень не найден")
		}
		var pretty any
		if err := json.Unmarshal(raw, &pretty); err != nil {
			return httpErrf(http.StatusBadRequest, "некорректный JSON теста: %v", err)
		}
		return writeJSONFile(snap, filepath.Join(dir, "quiz.json"), pretty)
	})
}

func (a *API) adminQuizDelete(w http.ResponseWriter, r *http.Request) {
	course, level := pathID(w, r, "course"), r.PathValue("level")
	if course == "" {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		if !validID(level) {
			return httpErrf(http.StatusBadRequest, "некорректный id уровня")
		}
		path := filepath.Join(a.levelPath(course, level), "quiz.json")
		if err := snap.rememberFile(path); err != nil {
			return err
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	})
}

func (a *API) adminTaskUpsert(w http.ResponseWriter, r *http.Request) {
	course, level := pathID(w, r, "course"), r.PathValue("level")
	if course == "" {
		return
	}
	var req adminTaskDTO
	if !readJSON(w, r, &req) {
		return
	}
	if id := r.PathValue("task"); id != "" { // PUT: id из пути
		req.ID = id
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		if !validID(level) || !validID(req.ID) {
			return httpErrf(http.StatusBadRequest, "некорректный id")
		}
		levelDir := a.levelPath(course, level)
		if _, err := os.Stat(levelDir); err != nil {
			return httpErrf(http.StatusNotFound, "уровень не найден")
		}
		dir := filepath.Join(levelDir, "tasks", req.ID)
		if err := writeJSONFile(snap, filepath.Join(dir, "task.json"), req.taskJSON); err != nil {
			return err
		}
		if err := writeTextFile(snap, filepath.Join(dir, "statement.md"), req.StatementMD); err != nil {
			return err
		}
		if err := writeTextFile(snap, filepath.Join(dir, "starter.go"), req.StarterCode); err != nil {
			return err
		}
		testPath := filepath.Join(dir, "main_test.go")
		if req.Type == content.TaskUnitTest {
			return writeTextFile(snap, testPath, req.TestFile)
		}
		// io-задача: main_test.go не должен оставаться от прежнего типа
		if err := snap.rememberFile(testPath); err != nil {
			return err
		}
		if err := os.Remove(testPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	})
}

func (a *API) adminTaskDelete(w http.ResponseWriter, r *http.Request) {
	course, level, task := pathID(w, r, "course"), r.PathValue("level"), r.PathValue("task")
	if course == "" {
		return
	}
	a.mutateContent(w, func(snap *fsSnapshot) error {
		if !validID(level) || !validID(task) {
			return httpErrf(http.StatusBadRequest, "некорректный id")
		}
		dir := filepath.Join(a.levelPath(course, level), "tasks", task)
		if _, err := os.Stat(dir); err != nil {
			return httpErrf(http.StatusNotFound, "задача не найдена")
		}
		return snap.moveDirToTrash(dir)
	})
}

// ── Картинки ────────────────────────────────────────────────────────────────

func (a *API) adminImageUpload(w http.ResponseWriter, r *http.Request) {
	course, level := pathID(w, r, "course"), r.PathValue("level")
	if course == "" {
		return
	}
	if !validID(level) {
		writeErr(w, http.StatusBadRequest, "некорректный id уровня")
		return
	}
	if !checkLocalOrigin(r) {
		writeErr(w, http.StatusForbidden, "загрузка разрешена только с этого сайта")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "нет файла в поле file: "+err.Error())
		return
	}
	defer file.Close()

	name := strings.ToLower(filepath.Base(header.Filename))
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if !imageNameRe.MatchString(name) || !imageExts[filepath.Ext(name)] || winReserved[base] {
		writeErr(w, http.StatusBadRequest, "имя файла: строчные латинские/цифры/точки/дефисы + расширение картинки")
		return
	}

	// Картинки не валидируются загрузчиком контента — снапшот-механика не нужна,
	// но пишем под fsMu, чтобы не гоняться с reload.
	a.fsMu.Lock()
	defer a.fsMu.Unlock()
	dir := filepath.Join(a.levelPath(course, level), "img")
	if _, err := os.Stat(a.levelPath(course, level)); err != nil {
		writeErr(w, http.StatusNotFound, "уровень не найден")
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	dst, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"name": name, "markdown": fmt.Sprintf("![описание](img/%s)", name)})
}

func (a *API) adminImageDelete(w http.ResponseWriter, r *http.Request) {
	course, level, name := pathID(w, r, "course"), r.PathValue("level"), strings.ToLower(r.PathValue("name"))
	if course == "" {
		return
	}
	if !validID(level) || !imageNameRe.MatchString(name) || !imageExts[filepath.Ext(name)] {
		writeErr(w, http.StatusBadRequest, "некорректное имя")
		return
	}
	a.fsMu.Lock()
	defer a.fsMu.Unlock()
	if err := os.Remove(filepath.Join(a.levelPath(course, level), "img", name)); err != nil {
		writeErr(w, http.StatusNotFound, "картинка не найдена")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
