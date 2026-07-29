package content

import (
	"fmt"
	"path/filepath"
	"strings"
)

// validateCourse проверяет граф уровней курса: ссылки requires и отсутствие циклов.
func validateCourse(c *Course) []error {
	var errs []error
	levelPath := func(id string) string {
		return filepath.Join(c.Dir, "levels", id, "level.json")
	}

	for _, id := range c.LevelOrder {
		level, ok := c.Levels[id]
		if !ok {
			continue // уровень не загрузился, ошибка уже есть
		}
		for _, req := range level.Requires {
			if req == id {
				errs = append(errs, fmt.Errorf("%s: уровень требует сам себя", levelPath(id)))
			} else if _, ok := c.Levels[req]; !ok {
				errs = append(errs, fmt.Errorf("%s: requires ссылается на несуществующий уровень %q", levelPath(id), req))
			}
		}
	}
	if len(errs) > 0 {
		return errs // сначала чиним ссылки, потом ищем циклы
	}

	// Поиск цикла: DFS с тремя цветами, восстанавливаем путь цикла для сообщения.
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[string]int{}
	var stack []string
	var cyclePath []string

	var visit func(id string) bool
	visit = func(id string) bool {
		color[id] = gray
		stack = append(stack, id)
		for _, req := range c.Levels[id].Requires {
			switch color[req] {
			case gray:
				start := 0
				for i, s := range stack {
					if s == req {
						start = i
						break
					}
				}
				cyclePath = append(append([]string{}, stack[start:]...), req)
				return true
			case white:
				if visit(req) {
					return true
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[id] = black
		return false
	}

	for _, id := range c.LevelOrder {
		if _, ok := c.Levels[id]; !ok {
			continue
		}
		if color[id] == white && visit(id) {
			errs = append(errs, fmt.Errorf("%s: цикл в requires: %s",
				filepath.Join(c.Dir, "course.json"), strings.Join(cyclePath, " -> ")))
			break
		}
	}
	return errs
}
