package store

import "testing"

func TestMemoryCourseFactsAll(t *testing.T) {
	m := NewMemory()

	// Пользователь 1: тест + задача по курсу go, плюс шум по другому курсу.
	if err := m.RecordQuizAttempt(1, "go", "l1", true); err != nil {
		t.Fatal(err)
	}
	if err := m.RecordTaskRun(1, "go", "l1", "t1", "code", true); err != nil {
		t.Fatal(err)
	}
	if err := m.RecordQuizAttempt(1, "other", "x1", true); err != nil {
		t.Fatal(err)
	}
	// Пользователь 2: только черновик.
	if err := m.SaveDraft(2, "go", "l1", "t1", "draft"); err != nil {
		t.Fatal(err)
	}

	byUser, err := m.CourseFactsAll("go")
	if err != nil {
		t.Fatal(err)
	}
	if len(byUser) != 2 {
		t.Fatalf("ожидались факты двух пользователей, получено %d", len(byUser))
	}

	f1 := byUser[1]
	if f1 == nil || !f1.Level("l1").QuizPassed {
		t.Fatalf("факты пользователя 1 неполны: %+v", f1)
	}
	if !f1.Task("l1", "t1").Completed {
		t.Fatal("задача пользователя 1 должна быть completed")
	}
	if _, ok := f1.Levels["x1"]; ok {
		t.Fatal("факты чужого курса подмешались в выборку")
	}

	f2 := byUser[2]
	if f2 == nil || f2.Task("l1", "t1").Draft != "draft" {
		t.Fatalf("черновик пользователя 2 потерян: %+v", f2)
	}
	if f2.Task("l1", "t1").Completed {
		t.Fatal("черновик не должен считаться выполнением")
	}
}
