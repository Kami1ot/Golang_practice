package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"gopractice/internal/content"
)

// RunUnitTest кладёт код пользователя (solution.go) рядом со скрытым
// авторским main_test.go и запускает `go test -json`.
// Ошибки компиляции и результаты каждого Test* разбираются из JSON-потока.
func (r *Runner) RunUnitTest(ctx context.Context, task *content.Task, code string) (*Result, error) {
	r.sem <- struct{}{}
	defer func() { <-r.sem }()

	ws, err := r.makeWorkspace(map[string]string{
		"solution.go":  code,
		"main_test.go": task.TestFile,
	})
	if err != nil {
		return nil, err
	}
	defer removeAllRetry(ws)

	timeout := time.Duration(task.Timeout()) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout+5*time.Second)
	defer cancel()

	stdout := &cappedWriter{limit: outputCap, onOverflow: cancel}
	stderr := &cappedWriter{limit: outputCap, onOverflow: cancel}

	cmd := exec.CommandContext(ctx, r.goBin, "test", "-json", "-count=1",
		"-timeout", timeout.String(), ".")
	cmd.Dir = ws
	cmd.Env = runEnv()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = waitDelay

	runErr := cmd.Run()

	if ctx.Err() != nil {
		return &Result{
			Status: StatusTimeout,
			Tests: []TestResult{{
				Name:   "go test",
				Output: "превышен лимит времени (" + timeout.String() + ")",
			}},
		}, nil
	}

	tests, pkgOutput := parseTestEvents(stdout.String())

	// Ни одного теста не запустилось, а команда упала — код не собрался.
	if len(tests) == 0 && runErr != nil {
		combined := strings.TrimSpace(pkgOutput + "\n" + stderr.String())
		return &Result{
			Status:        StatusCompileError,
			CompileOutput: cleanPaths(combined, ws),
		}, nil
	}

	result := &Result{Status: StatusPassed}
	for _, t := range tests {
		t.Output = cleanPaths(t.Output, ws)
		result.Tests = append(result.Tests, t)
		if !t.Passed {
			if strings.Contains(t.Output, "test timed out") {
				result.Status = StatusTimeout
			} else if result.Status == StatusPassed {
				result.Status = StatusFailed
			}
		}
	}
	return result, nil
}

// testEvent — строка вывода go test -json (см. `go doc test2json`).
type testEvent struct {
	Action  string  `json:"Action"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

// parseTestEvents собирает результаты по каждому Test* и вывод уровня пакета
// (туда попадают ошибки компиляции).
func parseTestEvents(raw string) ([]TestResult, string) {
	type acc struct {
		output  []string
		passed  bool
		done    bool
		elapsed float64
	}
	byTest := map[string]*acc{}
	var order []string
	var pkgOutput strings.Builder

	sc := bufio.NewScanner(strings.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		var ev testEvent
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			continue // построчный мусор (например, усечение capped-writer)
		}
		if ev.Test == "" {
			if ev.Action == "output" {
				pkgOutput.WriteString(ev.Output)
			}
			continue
		}
		a := byTest[ev.Test]
		if a == nil {
			a = &acc{}
			byTest[ev.Test] = a
			order = append(order, ev.Test)
		}
		switch ev.Action {
		case "output":
			line := ev.Output
			trimmed := strings.TrimSpace(line)
			// Служебные строки фреймворка не несут информации для пользователя.
			if strings.HasPrefix(trimmed, "=== ") || strings.HasPrefix(trimmed, "--- ") {
				break
			}
			a.output = append(a.output, line)
		case "pass":
			a.passed, a.done, a.elapsed = true, true, ev.Elapsed
		case "fail":
			a.done, a.elapsed = true, ev.Elapsed
		}
	}

	var tests []TestResult
	for _, name := range order {
		a := byTest[name]
		if !a.done {
			continue // подтесты в процессе и т.п.
		}
		tests = append(tests, TestResult{
			Name:       name,
			Passed:     a.passed,
			Output:     strings.TrimSpace(strings.Join(a.output, "")),
			DurationMs: int64(a.elapsed * 1000),
		})
	}
	return tests, pkgOutput.String()
}

// Run выбирает механизм проверки по типу задачи.
func (r *Runner) Run(ctx context.Context, task *content.Task, code string) (*Result, error) {
	if task.Type == content.TaskUnitTest {
		return r.RunUnitTest(ctx, task, code)
	}
	return r.RunIO(ctx, task, code)
}
