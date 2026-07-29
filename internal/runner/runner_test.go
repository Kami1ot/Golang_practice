package runner

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"gopractice/internal/content"
)

// Тесты запускают настоящий go-тулчейн: медленнее юнит-тестов,
// но проверяют именно то, что увидит пользователь.

func newRunner(t *testing.T) *Runner {
	t.Helper()
	r, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return r
}

func ioTask(timeoutSec int, tests ...content.TestCase) *content.Task {
	return &content.Task{ID: "t", Type: content.TaskIO, Title: "t", TimeoutSec: timeoutSec, Tests: tests}
}

func TestIOPass(t *testing.T) {
	r := newRunner(t)
	task := ioTask(10, content.TestCase{Name: "сумма", Stdin: "3 4\n", Stdout: "7\n"})
	res, err := r.RunIO(context.Background(), task, `package main

import "fmt"

func main() {
	var a, b int
	fmt.Scan(&a, &b)
	fmt.Println(a + b)
}`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, тесты: %+v", res.Status, res.Tests)
	}
}

func TestIOWrongOutput(t *testing.T) {
	r := newRunner(t)
	task := ioTask(10, content.TestCase{Name: "точный вывод", Stdin: "", Stdout: "да\n"})
	res, err := r.RunIO(context.Background(), task, `package main

import "fmt"

func main() { fmt.Println("нет") }`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, ожидался failed", res.Status)
	}
	tr := res.Tests[0]
	if tr.Expected != "да" || tr.Actual != "нет" {
		t.Fatalf("diff не заполнен: expected=%q actual=%q", tr.Expected, tr.Actual)
	}
}

func TestCompileErrorCleanPaths(t *testing.T) {
	r := newRunner(t)
	task := ioTask(10, content.TestCase{Name: "x", Stdout: "x"})
	res, err := r.RunIO(context.Background(), task, `package main

func main() { fmt.Println("забыл import") }`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusCompileError {
		t.Fatalf("status = %s, ожидался compile_error", res.Status)
	}
	if !strings.Contains(res.CompileOutput, "fmt") {
		t.Fatalf("нет сообщения компилятора: %q", res.CompileOutput)
	}
	if strings.Contains(res.CompileOutput, "gotask-") {
		t.Fatalf("путь временной папки не вычищен: %q", res.CompileOutput)
	}
}

func TestTimeoutKillsProcess(t *testing.T) {
	r := newRunner(t)
	task := ioTask(2,
		content.TestCase{Name: "вечный цикл", Stdout: "x"},
		content.TestCase{Name: "после таймаута", Stdout: "x"},
	)
	start := time.Now()
	res, err := r.RunIO(context.Background(), task, `package main

func main() { for {} }`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusTimeout {
		t.Fatalf("status = %s, ожидался timeout", res.Status)
	}
	if elapsed := time.Since(start); elapsed > 15*time.Second {
		t.Fatalf("убийство процесса заняло %s", elapsed)
	}
	if !res.Tests[1].Skipped {
		t.Fatal("второй тест должен быть пропущен после таймаута")
	}
}

func TestOutputFloodCapped(t *testing.T) {
	r := newRunner(t)
	task := ioTask(10, content.TestCase{Name: "флуд", Stdout: "x"})
	res, err := r.RunIO(context.Background(), task, `package main

import "fmt"

func main() {
	for {
		fmt.Println("СПАМ СПАМ СПАМ СПАМ СПАМ СПАМ СПАМ СПАМ")
	}
}`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusTimeout {
		t.Fatalf("status = %s, ожидался timeout (переполнение вывода)", res.Status)
	}
	if len(res.Tests[0].Actual) > outputCap+1024 {
		t.Fatalf("вывод не ограничен: %d байт", len(res.Tests[0].Actual))
	}
}

func TestRuntimeError(t *testing.T) {
	r := newRunner(t)
	task := ioTask(10, content.TestCase{Name: "паника", Stdout: "x"})
	res, err := r.RunIO(context.Background(), task, `package main

func main() {
	var s []int
	_ = s[5]
}`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusRuntimeError {
		t.Fatalf("status = %s, ожидался runtime_error", res.Status)
	}
	if !strings.Contains(res.Tests[0].Output, "index out of range") {
		t.Fatalf("нет текста паники: %q", res.Tests[0].Output)
	}
}

func TestUnicodeOutput(t *testing.T) {
	r := newRunner(t)
	task := ioTask(10, content.TestCase{Name: "кириллица", Stdout: "Привет, Go!\n"})
	res, err := r.RunIO(context.Background(), task, `package main

import "fmt"

func main() { fmt.Println("Привет, Go!") }`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusPassed {
		t.Fatalf("status = %s: %+v", res.Status, res.Tests)
	}
}

const sumTestFile = `package main

import "testing"

func TestSum(t *testing.T) {
	if got := Sum(3, 4); got != 7 {
		t.Errorf("Sum(3, 4) = %d, ожидалось 7", got)
	}
}

func TestSumNegative(t *testing.T) {
	if got := Sum(-2, -3); got != -5 {
		t.Errorf("Sum(-2, -3) = %d, ожидалось -5", got)
	}
}
`

func unitTask() *content.Task {
	return &content.Task{ID: "t", Type: content.TaskUnitTest, Title: "t", TimeoutSec: 15, TestFile: sumTestFile}
}

func TestUnitTestPass(t *testing.T) {
	r := newRunner(t)
	res, err := r.RunUnitTest(context.Background(), unitTask(), `package main

func Sum(a, b int) int { return a + b }`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusPassed {
		t.Fatalf("status = %s: %+v", res.Status, res.Tests)
	}
	if len(res.Tests) != 2 {
		t.Fatalf("тестов %d, ожидалось 2", len(res.Tests))
	}
}

func TestUnitTestFailWithNames(t *testing.T) {
	r := newRunner(t)
	res, err := r.RunUnitTest(context.Background(), unitTask(), `package main

func Sum(a, b int) int { return a - b }`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, ожидался failed", res.Status)
	}
	var failedName string
	for _, tr := range res.Tests {
		if !tr.Passed {
			failedName = tr.Name
			if !strings.Contains(tr.Output, "ожидалось") {
				t.Fatalf("нет сообщения t.Errorf: %q", tr.Output)
			}
		}
	}
	if failedName == "" {
		t.Fatal("ни один тест не помечен как проваленный")
	}
}

func TestUnitTestCompileError(t *testing.T) {
	r := newRunner(t)
	res, err := r.RunUnitTest(context.Background(), unitTask(), `package main

func Sum(a, b int) int { return нет_такой_переменной }`)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != StatusCompileError {
		t.Fatalf("status = %s, ожидался compile_error", res.Status)
	}
	if strings.Contains(res.CompileOutput, "gotask-") {
		t.Fatalf("путь временной папки не вычищен: %q", res.CompileOutput)
	}
}

// Горутины не должны утекать даже после таймаутов.
func TestNoGoroutineLeak(t *testing.T) {
	r := newRunner(t)
	task := ioTask(2, content.TestCase{Name: "вечный цикл", Stdout: "x"})
	before := runtime.NumGoroutine()
	for i := 0; i < 3; i++ {
		if _, err := r.RunIO(context.Background(), task, `package main

func main() { select {} }`); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(3 * time.Second) // WaitDelay и внутренние копировщики os/exec должны завершиться
	after := runtime.NumGoroutine()
	if after > before+2 {
		t.Fatalf("горутины утекли: было %d, стало %d", before, after)
	}
}

func TestNormalize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a\r\nb\r\n", "a\nb"},
		{"a  \nb\t\n\n\n", "a\nb"},
		{"7\n", "7"},
		{"", ""},
		{"  ведущие пробелы важны", "  ведущие пробелы важны"},
	}
	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, ожидалось %q", c.in, got, c.want)
		}
	}
}
