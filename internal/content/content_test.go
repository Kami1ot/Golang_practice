package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTree разворачивает карту относительный-путь -> содержимое во временную папку.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// validCourse — минимальный валидный курс из двух уровней.
func validCourse() map[string]string {
	return map[string]string{
		"courses/demo/course.json": `{"id":"demo","title":"Демо","description":"d","levels":["l1","l2"]}`,
		"courses/demo/levels/l1/level.json": `{"id":"l1","title":"Первый","summary":"s","requires":[],"xp":10}`,
		"courses/demo/levels/l1/lesson.md":  "# Урок 1",
		"courses/demo/levels/l1/quiz.json": `{"passScore":1,"questions":[
			{"id":"q1","type":"single","prompt":"?","options":["a","b"],"answer":0,"explanation":"e"}]}`,
		"courses/demo/levels/l1/tasks/t1/task.json": `{"id":"t1","type":"io","title":"Задача","order":1,
			"tests":[{"name":"базовый","stdin":"","stdout":"ok\n"}]}`,
		"courses/demo/levels/l1/tasks/t1/statement.md": "Сделайте.",
		"courses/demo/levels/l1/tasks/t1/starter.go":   "package main\nfunc main() {}\n",
		"courses/demo/levels/l2/level.json": `{"id":"l2","title":"Второй","summary":"s","requires":["l1"],"xp":10}`,
		"courses/demo/levels/l2/lesson.md":  "# Урок 2",
	}
}

func TestLoadValid(t *testing.T) {
	cat, err := Load(writeTree(t, validCourse()))
	if err != nil {
		t.Fatalf("валидный курс не загрузился: %v", err)
	}
	course := cat.Courses["demo"]
	if course == nil {
		t.Fatal("курс demo не найден в каталоге")
	}
	if len(course.Levels) != 2 {
		t.Fatalf("уровней %d, ожидалось 2", len(course.Levels))
	}
	l1 := course.Levels["l1"]
	if l1.Quiz == nil || len(l1.Quiz.Questions) != 1 {
		t.Fatal("quiz уровня l1 не загрузился")
	}
	if got := l1.Quiz.Questions[0].AnswerIndex; got != 0 {
		t.Fatalf("AnswerIndex = %d, ожидалось 0", got)
	}
	if len(l1.Tasks) != 1 || l1.Tasks[0].StarterCode == "" {
		t.Fatal("задача уровня l1 не загрузилась")
	}
	if l2 := course.Levels["l2"]; l2.Quiz != nil || len(l2.Tasks) != 0 {
		t.Fatal("у l2 не должно быть теста и задач")
	}
}

// битые фикстуры: каждая мутация должна дать ошибку с нужным текстом
func TestLoadErrors(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(files map[string]string)
		wantSub string
	}{
		{
			"цикл в requires",
			func(f map[string]string) {
				f["courses/demo/levels/l1/level.json"] = `{"id":"l1","title":"Первый","summary":"s","requires":["l2"],"xp":10}`
			},
			"цикл в requires",
		},
		{
			"битая ссылка requires",
			func(f map[string]string) {
				f["courses/demo/levels/l2/level.json"] = `{"id":"l2","title":"Второй","summary":"s","requires":["нет-такого"],"xp":10}`
			},
			"несуществующий уровень",
		},
		{
			"id не совпадает с папкой",
			func(f map[string]string) {
				f["courses/demo/levels/l2/level.json"] = `{"id":"другой","title":"Второй","summary":"s","requires":["l1"],"xp":10}`
			},
			"не совпадает с именем папки",
		},
		{
			"нет lesson.md",
			func(f map[string]string) { delete(f, "courses/demo/levels/l2/lesson.md") },
			"lesson.md",
		},
		{
			"папка уровня не в course.json",
			func(f map[string]string) {
				f["courses/demo/levels/l3/level.json"] = `{"id":"l3","title":"X","summary":"s","requires":[]}`
				f["courses/demo/levels/l3/lesson.md"] = "# X"
			},
			"не указана в course.json",
		},
		{
			"неизвестный тип вопроса",
			func(f map[string]string) {
				f["courses/demo/levels/l1/quiz.json"] = `{"passScore":1,"questions":[
					{"id":"q1","type":"странный","prompt":"?","answer":0}]}`
			},
			"неизвестный type",
		},
		{
			"answer вне диапазона options",
			func(f map[string]string) {
				f["courses/demo/levels/l1/quiz.json"] = `{"passScore":1,"questions":[
					{"id":"q1","type":"single","prompt":"?","options":["a","b"],"answer":5}]}`
			},
			"вне диапазона",
		},
		{
			"io-задача без тестов",
			func(f map[string]string) {
				f["courses/demo/levels/l1/tasks/t1/task.json"] = `{"id":"t1","type":"io","title":"Задача","order":1}`
			},
			"без tests",
		},
		{
			"unittest без main_test.go",
			func(f map[string]string) {
				f["courses/demo/levels/l1/tasks/t1/task.json"] = `{"id":"t1","type":"unittest","title":"Задача","order":1}`
			},
			"без main_test.go",
		},
		{
			"синтаксическая ошибка JSON со строкой и колонкой",
			func(f map[string]string) {
				f["courses/demo/course.json"] = "{\n  \"id\": \"demo\",\n  оп\n}"
			},
			"course.json:3:",
		},
		{
			"неизвестное поле JSON",
			func(f map[string]string) {
				f["courses/demo/levels/l2/level.json"] = `{"id":"l2","title":"Второй","summary":"s","requires":["l1"],"опечатка":1}`
			},
			"unknown field",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := validCourse()
			tc.mutate(files)
			_, err := Load(writeTree(t, files))
			if err == nil {
				t.Fatal("ожидалась ошибка, получен nil")
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("в ошибке нет %q:\n%v", tc.wantSub, err)
			}
		})
	}
}

// Пустой курс и пустой каталог допустимы: их создаёт админка до наполнения.
func TestEmptyCourseAndCatalogAllowed(t *testing.T) {
	cat, err := Load(writeTree(t, map[string]string{
		"courses/empty/course.json": `{"id":"empty","title":"Пустой","description":"d","levels":[]}`,
	}))
	if err != nil {
		t.Fatalf("пустой курс должен загружаться: %v", err)
	}
	if len(cat.Courses["empty"].Levels) != 0 {
		t.Fatal("у пустого курса не должно быть уровней")
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "courses"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err != nil {
		t.Fatalf("пустой каталог должен загружаться: %v", err)
	}
}

func TestHintsValidation(t *testing.T) {
	files := validCourse()
	files["courses/demo/levels/l1/tasks/t1/task.json"] = `{"id":"t1","type":"io","title":"Задача","order":1,
		"hints":["нормальная", "  "],
		"tests":[{"name":"базовый","stdin":"","stdout":"ok\n"}]}`
	_, err := Load(writeTree(t, files))
	if err == nil || !strings.Contains(err.Error(), "пустая подсказка") {
		t.Fatalf("пустая подсказка не поймана: %v", err)
	}

	files["courses/demo/levels/l1/tasks/t1/task.json"] = `{"id":"t1","type":"io","title":"Задача","order":1,
		"hints":["первая", "вторая"],
		"tests":[{"name":"базовый","stdin":"","stdout":"ok\n"}]}`
	cat, err := Load(writeTree(t, files))
	if err != nil {
		t.Fatalf("валидные hints не прошли: %v", err)
	}
	if hints := cat.Courses["demo"].Levels["l1"].Tasks[0].Hints; len(hints) != 2 {
		t.Fatalf("hints не загрузились: %v", hints)
	}
}

func TestTagsValidation(t *testing.T) {
	files := validCourse()
	files["courses/demo/course.json"] = `{"id":"demo","title":"Демо","description":"d","tags":["go","  "],"levels":["l1","l2"]}`
	_, err := Load(writeTree(t, files))
	if err == nil || !strings.Contains(err.Error(), "пустой тег") {
		t.Fatalf("пустой тег не пойман: %v", err)
	}

	files["courses/demo/course.json"] = `{"id":"demo","title":"Демо","description":"d","tags":["go","GO"],"levels":["l1","l2"]}`
	_, err = Load(writeTree(t, files))
	if err == nil || !strings.Contains(err.Error(), "дублируется") {
		t.Fatalf("дубликат тега (без учёта регистра) не пойман: %v", err)
	}

	files["courses/demo/course.json"] = `{"id":"demo","title":"Демо","description":"d","tags":["go","основы"],"levels":["l1","l2"]}`
	cat, err := Load(writeTree(t, files))
	if err != nil {
		t.Fatalf("валидные tags не прошли: %v", err)
	}
	if tags := cat.Courses["demo"].Tags; len(tags) != 2 || tags[0] != "go" || tags[1] != "основы" {
		t.Fatalf("tags не загрузились: %v", tags)
	}
}

func TestSelfRequireIsCycle(t *testing.T) {
	files := validCourse()
	files["courses/demo/levels/l1/level.json"] = `{"id":"l1","title":"Первый","summary":"s","requires":["l1"],"xp":10}`
	_, err := Load(writeTree(t, files))
	if err == nil || !strings.Contains(err.Error(), "сам себя") {
		t.Fatalf("самозависимость не поймана: %v", err)
	}
}
