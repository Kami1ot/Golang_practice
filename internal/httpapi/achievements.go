package httpapi

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"gopractice/internal/content"
	"gopractice/internal/store"
)

// Ачивки: условия — код поверх фактов (в БД хранится только факт выдачи).

type achievementDef struct {
	ID          string
	Title       string
	Description string
	Icon        string
	// Condition оценивается по агрегату всех курсов пользователя.
	Condition func(s *userStats) bool
}

type achievementDTO struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	GrantedAt   *time.Time `json:"grantedAt,omitempty"`
}

// userStats — агрегированные факты пользователя по всем курсам текущего каталога.
type userStats struct {
	completedLevels   int
	completedBonus    int
	completedCourses  []string // названия завершённых курсов
	xp                int
	tasksDone         int
	quizFirstTry      bool // хоть один тест сдан с первой попытки
	taskFirstTry      bool // хоть одна задача решена с первой попытки
	taskAfterMany     bool // хоть одна задача решена с ≥5 попытки
	nightOwl          bool // задача решена в 00:00–05:00
	quizzesPassed     int
}

func collectStats(cat *content.Catalog, st store.Store, userID int64) (*userStats, error) {
	s := &userStats{}
	for _, courseID := range cat.Order {
		c := cat.Courses[courseID]
		facts, err := st.CourseFacts(userID, courseID)
		if err != nil {
			return nil, err
		}
		done := completedSet(c, facts)
		allRequiredDone := true
		for _, l := range c.Levels {
			if done[l.ID] {
				s.completedLevels++
				s.xp += l.XP
				if l.Optional {
					s.completedBonus++
				}
			} else if !l.Optional {
				allRequiredDone = false
			}
			lf := facts.Level(l.ID)
			if lf.QuizPassed {
				s.quizzesPassed++
				if lf.QuizAttempts == 1 {
					s.quizFirstTry = true
				}
			}
			for _, t := range l.Tasks {
				tf := facts.Task(l.ID, t.ID)
				if tf.Completed {
					s.tasksDone++
					if tf.Attempts == 1 {
						s.taskFirstTry = true
					}
					if tf.Attempts >= 5 {
						s.taskAfterMany = true
					}
					if h := tf.CompletedAt.Hour(); !tf.CompletedAt.IsZero() && h < 5 {
						s.nightOwl = true
					}
				}
			}
		}
		if allRequiredDone && len(c.Levels) > 0 {
			s.completedCourses = append(s.completedCourses, c.Title)
		}
	}
	return s, nil
}

// achievementDefs строится по текущему каталогу (per-course ачивки — динамические).
func achievementDefs(cat *content.Catalog) []achievementDef {
	defs := []achievementDef{
		{"first-steps", "Первые шаги", "Пройдите свой первый уровень", "🐣",
			func(s *userStats) bool { return s.completedLevels >= 1 }},
		{"quiz-perfect", "Перфекционист", "Сдайте тест с первой попытки", "🎯",
			func(s *userStats) bool { return s.quizFirstTry }},
		{"task-first-try", "С первого раза", "Решите задачу с первой попытки", "⚡",
			func(s *userStats) bool { return s.taskFirstTry }},
		{"persistence", "Упорство", "Решите задачу с пятой попытки или позже", "🪨",
			func(s *userStats) bool { return s.taskAfterMany }},
		{"bonus-hunter", "Бонус-хантер", "Пройдите бонусный уровень", "⭐",
			func(s *userStats) bool { return s.completedBonus >= 1 }},
		{"xp-100", "Сотня", "Наберите 100 XP", "💯",
			func(s *userStats) bool { return s.xp >= 100 }},
		{"xp-250", "Четверть тысячи", "Наберите 250 XP", "🚀",
			func(s *userStats) bool { return s.xp >= 250 }},
		{"xp-500", "Полтысячи", "Наберите 500 XP", "👑",
			func(s *userStats) bool { return s.xp >= 500 }},
		{"task-master", "Мастер задач", "Решите 10 задач", "🛠",
			func(s *userStats) bool { return s.tasksDone >= 10 }},
		{"night-owl", "Полуночник", "Решите задачу между полуночью и пятью утра", "🦉",
			func(s *userStats) bool { return s.nightOwl }},
	}
	for _, courseID := range cat.Order {
		c := cat.Courses[courseID]
		title := c.Title
		defs = append(defs, achievementDef{
			ID:          "course-" + c.ID,
			Title:       "Курс завершён: " + title,
			Description: fmt.Sprintf("Пройдите все обязательные уровни курса «%s»", title),
			Icon:        "🏆",
			Condition: func(s *userStats) bool {
				for _, done := range s.completedCourses {
					if done == title {
						return true
					}
				}
				return false
			},
		})
	}
	return defs
}

// evaluateAchievements проверяет условия и выдаёт новые ачивки.
// Возвращает ТОЛЬКО выданные этим вызовом (для баннера «Новая ачивка!»).
func (a *API) evaluateAchievements(u *store.User) []achievementDTO {
	cat := a.cat()
	stats, err := collectStats(cat, a.store, u.ID)
	if err != nil {
		log.Printf("ачивки: сбор статистики: %v", err)
		return nil
	}
	var fresh []achievementDTO
	for _, def := range achievementDefs(cat) {
		if !def.Condition(stats) {
			continue
		}
		granted, err := a.store.GrantAchievement(u.ID, def.ID)
		if err != nil {
			log.Printf("ачивки: выдача %s: %v", def.ID, err)
			continue
		}
		if granted {
			now := time.Now()
			fresh = append(fresh, achievementDTO{
				ID: def.ID, Title: def.Title, Description: def.Description,
				Icon: def.Icon, GrantedAt: &now,
			})
		}
	}
	return fresh
}

type profileDTO struct {
	User         userDTO          `json:"user"`
	RegisteredAt time.Time        `json:"registeredAt"`
	Stats        profileStatsDTO  `json:"stats"`
	Achievements []achievementDTO `json:"achievements"`
}

type profileStatsDTO struct {
	XP              int `json:"xp"`
	LevelsCompleted int `json:"levelsCompleted"`
	TasksDone       int `json:"tasksDone"`
	QuizzesPassed   int `json:"quizzesPassed"`
}

func (a *API) handleProfile(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	// Бэкфилл: прогресс мог появиться до введения ачивок.
	a.evaluateAchievements(u)

	cat := a.cat()
	stats, err := collectStats(cat, a.store, u.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	granted, err := a.store.Achievements(u.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}

	defs := achievementDefs(cat)
	list := make([]achievementDTO, 0, len(defs))
	for _, def := range defs {
		dto := achievementDTO{ID: def.ID, Title: def.Title, Description: def.Description, Icon: def.Icon}
		if at, ok := granted[def.ID]; ok {
			t := at
			dto.GrantedAt = &t
		}
		list = append(list, dto)
	}
	// Полученные — вперёд, по свежести; затем залоченные.
	sort.SliceStable(list, func(i, j int) bool {
		gi, gj := list[i].GrantedAt != nil, list[j].GrantedAt != nil
		if gi != gj {
			return gi
		}
		if gi && gj {
			return list[i].GrantedAt.After(*list[j].GrantedAt)
		}
		return false
	})

	writeJSON(w, http.StatusOK, profileDTO{
		User:         a.userDTO(u),
		RegisteredAt: u.CreatedAt,
		Stats: profileStatsDTO{
			XP:              stats.xp,
			LevelsCompleted: stats.completedLevels,
			TasksDone:       stats.tasksDone,
			QuizzesPassed:   stats.quizzesPassed,
		},
		Achievements: list,
	})
}
