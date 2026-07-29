package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Load читает и валидирует все курсы из dir (обычно "content").
// При любой ошибке возвращает nil и объединённый список проблем с путями файлов.
func Load(dir string) (*Catalog, error) {
	coursesDir := filepath.Join(dir, "courses")
	entries, err := os.ReadDir(coursesDir)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать %s: %w", coursesDir, err)
	}

	cat := &Catalog{Courses: map[string]*Course{}}
	var errs []error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		course, cerrs := loadCourse(filepath.Join(coursesDir, e.Name()), e.Name())
		if len(cerrs) > 0 {
			errs = append(errs, cerrs...)
			continue
		}
		cat.Courses[course.ID] = course
		cat.Order = append(cat.Order, course.ID)
	}
	sort.Strings(cat.Order)

	// Пустой каталог допустим: админ мог ещё не создать ни одного курса.
	if len(errs) == 0 && len(cat.Courses) == 0 {
		log.Printf("внимание: в %s нет ни одного курса", coursesDir)
	}
	for _, c := range cat.Courses {
		errs = append(errs, validateCourse(c)...)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return cat, nil
}

func loadCourse(dir, dirName string) (*Course, []error) {
	var errs []error
	course := &Course{Dir: dir, Levels: map[string]*Level{}}

	coursePath := filepath.Join(dir, "course.json")
	if err := decodeJSONFile(coursePath, course); err != nil {
		return nil, []error{err}
	}
	if course.ID != dirName {
		errs = append(errs, fmt.Errorf("%s: id %q не совпадает с именем папки %q", coursePath, course.ID, dirName))
	}
	if course.Title == "" {
		errs = append(errs, fmt.Errorf("%s: пустой title", coursePath))
	}

	seenTag := map[string]bool{}
	for i, tag := range course.Tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			errs = append(errs, fmt.Errorf("%s: tags[%d]: пустой тег", coursePath, i))
			continue
		}
		key := strings.ToLower(trimmed)
		if seenTag[key] {
			errs = append(errs, fmt.Errorf("%s: тег %q дублируется", coursePath, tag))
		}
		seenTag[key] = true
	}
	// Пустой levels допустим: курс только что создан в админке.

	seen := map[string]bool{}
	for _, levelID := range course.LevelOrder {
		if seen[levelID] {
			errs = append(errs, fmt.Errorf("%s: уровень %q указан дважды", coursePath, levelID))
			continue
		}
		seen[levelID] = true
		level, lerrs := loadLevel(filepath.Join(dir, "levels", levelID), levelID)
		if len(lerrs) > 0 {
			errs = append(errs, lerrs...)
			continue
		}
		course.Levels[level.ID] = level
	}

	// Папки уровней, забытые в course.json — почти наверняка ошибка автора.
	levelsDir := filepath.Join(dir, "levels")
	if entries, err := os.ReadDir(levelsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() && !seen[e.Name()] {
				errs = append(errs, fmt.Errorf("%s: папка уровня %q не указана в course.json", levelsDir, e.Name()))
			}
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}
	return course, nil
}

func loadLevel(dir, dirName string) (*Level, []error) {
	var errs []error
	level := &Level{Dir: dir}

	levelPath := filepath.Join(dir, "level.json")
	if err := decodeJSONFile(levelPath, level); err != nil {
		return nil, []error{err}
	}
	if level.ID != dirName {
		errs = append(errs, fmt.Errorf("%s: id %q не совпадает с именем папки %q", levelPath, level.ID, dirName))
	}
	if level.Title == "" {
		errs = append(errs, fmt.Errorf("%s: пустой title", levelPath))
	}

	lessonPath := filepath.Join(dir, "lesson.md")
	lesson, err := os.ReadFile(lessonPath)
	if err != nil {
		errs = append(errs, fmt.Errorf("%s: нет обязательного lesson.md", dir))
	} else {
		level.LessonMD = string(lesson)
	}

	quizPath := filepath.Join(dir, "quiz.json")
	if _, err := os.Stat(quizPath); err == nil {
		quiz := &Quiz{}
		if err := decodeJSONFile(quizPath, quiz); err != nil {
			errs = append(errs, err)
		} else {
			errs = append(errs, resolveQuiz(quiz, dir, quizPath)...)
			level.Quiz = quiz
		}
	}

	tasksDir := filepath.Join(dir, "tasks")
	if entries, err := os.ReadDir(tasksDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			task, terrs := loadTask(filepath.Join(tasksDir, e.Name()), e.Name())
			if len(terrs) > 0 {
				errs = append(errs, terrs...)
				continue
			}
			level.Tasks = append(level.Tasks, task)
		}
		sort.Slice(level.Tasks, func(i, j int) bool {
			if level.Tasks[i].Order != level.Tasks[j].Order {
				return level.Tasks[i].Order < level.Tasks[j].Order
			}
			return level.Tasks[i].ID < level.Tasks[j].ID
		})
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return level, nil
}

// resolveQuiz дозаполняет типизированные ответы вопросов и валидирует их структуру.
func resolveQuiz(quiz *Quiz, levelDir, quizPath string) []error {
	var errs []error
	fail := func(q *Question, format string, args ...any) {
		errs = append(errs, fmt.Errorf("%s: вопрос %q: %s", quizPath, q.ID, fmt.Sprintf(format, args...)))
	}

	if quiz.PassScore <= 0 || quiz.PassScore > 1 {
		quiz.PassScore = 1.0
	}
	if len(quiz.Questions) == 0 {
		return []error{fmt.Errorf("%s: пустой список questions", quizPath)}
	}

	seen := map[string]bool{}
	for _, q := range quiz.Questions {
		if q.ID == "" {
			fail(q, "пустой id")
			continue
		}
		if seen[q.ID] {
			fail(q, "id повторяется")
			continue
		}
		seen[q.ID] = true
		if q.Prompt == "" {
			fail(q, "пустой prompt")
		}
		if q.CodeFile != "" {
			code, err := os.ReadFile(filepath.Join(levelDir, q.CodeFile))
			if err != nil {
				fail(q, "codeFile %q не читается: %v", q.CodeFile, err)
			} else {
				q.Code = string(code)
			}
		}

		switch q.Type {
		case QSingle:
			if len(q.Options) < 2 {
				fail(q, "нужно минимум 2 options")
			}
			if err := json.Unmarshal(q.RawAnswer, &q.AnswerIndex); err != nil {
				fail(q, "answer должен быть индексом варианта: %v", err)
			} else if q.AnswerIndex < 0 || q.AnswerIndex >= len(q.Options) {
				fail(q, "answer %d вне диапазона options", q.AnswerIndex)
			}
		case QMulti:
			if len(q.Options) < 2 {
				fail(q, "нужно минимум 2 options")
			}
			if err := json.Unmarshal(q.RawAnswers, &q.AnswerIndices); err != nil {
				fail(q, "answers должен быть массивом индексов: %v", err)
			} else if len(q.AnswerIndices) == 0 {
				fail(q, "answers пуст")
			} else {
				sort.Ints(q.AnswerIndices)
				for _, idx := range q.AnswerIndices {
					if idx < 0 || idx >= len(q.Options) {
						fail(q, "индекс %d вне диапазона options", idx)
					}
				}
			}
		case QOutput:
			if q.Code == "" {
				fail(q, "нужен code или codeFile")
			}
			if err := json.Unmarshal(q.RawAnswer, &q.AnswerText); err != nil || q.AnswerText == "" {
				fail(q, "answer должен быть непустой строкой")
			}
		case QBlank:
			if err := json.Unmarshal(q.RawAnswers, &q.AnswerTexts); err != nil || len(q.AnswerTexts) == 0 {
				fail(q, "answers должен быть непустым массивом строк")
			}
		default:
			fail(q, "неизвестный type %q (ожидается single|multi|output|blank)", q.Type)
		}
	}
	return errs
}

func loadTask(dir, dirName string) (*Task, []error) {
	var errs []error
	task := &Task{Dir: dir}

	taskPath := filepath.Join(dir, "task.json")
	if err := decodeJSONFile(taskPath, task); err != nil {
		return nil, []error{err}
	}
	if task.ID != dirName {
		errs = append(errs, fmt.Errorf("%s: id %q не совпадает с именем папки %q", taskPath, task.ID, dirName))
	}
	if task.Title == "" {
		errs = append(errs, fmt.Errorf("%s: пустой title", taskPath))
	}

	statement, err := os.ReadFile(filepath.Join(dir, "statement.md"))
	if err != nil {
		errs = append(errs, fmt.Errorf("%s: нет обязательного statement.md", dir))
	} else {
		task.StatementMD = string(statement)
	}
	starter, err := os.ReadFile(filepath.Join(dir, "starter.go"))
	if err != nil {
		errs = append(errs, fmt.Errorf("%s: нет обязательного starter.go", dir))
	} else {
		task.StarterCode = string(starter)
	}

	for i, h := range task.Hints {
		if strings.TrimSpace(h) == "" {
			errs = append(errs, fmt.Errorf("%s: hints[%d]: пустая подсказка", taskPath, i))
		}
	}

	switch task.Type {
	case TaskIO:
		if len(task.Tests) == 0 {
			errs = append(errs, fmt.Errorf("%s: io-задача без tests", taskPath))
		}
		for i, tc := range task.Tests {
			if tc.Name == "" {
				errs = append(errs, fmt.Errorf("%s: tests[%d]: пустой name", taskPath, i))
			}
		}
	case TaskUnitTest:
		testFile, err := os.ReadFile(filepath.Join(dir, "main_test.go"))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: unittest-задача без main_test.go", dir))
		} else {
			task.TestFile = string(testFile)
		}
		if len(task.Tests) > 0 {
			errs = append(errs, fmt.Errorf("%s: у unittest-задачи не должно быть tests в task.json", taskPath))
		}
	default:
		errs = append(errs, fmt.Errorf("%s: неизвестный type %q (ожидается io|unittest)", taskPath, task.Type))
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return task, nil
}

// decodeJSONFile декодирует JSON-файл строго (без неизвестных полей),
// добавляя к ошибкам путь файла и строку:колонку.
func decodeJSONFile(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var syn *json.SyntaxError
		var typ *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syn):
			line, col := lineCol(data, syn.Offset)
			return fmt.Errorf("%s:%d:%d: %v", path, line, col, err)
		case errors.As(err, &typ):
			line, col := lineCol(data, typ.Offset)
			return fmt.Errorf("%s:%d:%d: поле %q: %v", path, line, col, typ.Field, err)
		default:
			return fmt.Errorf("%s: %v", path, err)
		}
	}
	return nil
}

func lineCol(data []byte, offset int64) (line, col int) {
	if offset > int64(len(data)) {
		offset = int64(len(data))
	}
	prefix := data[:offset]
	line = 1 + strings.Count(string(prefix), "\n")
	if i := bytes.LastIndexByte(prefix, '\n'); i >= 0 {
		col = int(offset) - i
	} else {
		col = int(offset) + 1
	}
	return line, col
}
