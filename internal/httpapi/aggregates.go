package httpapi

import (
	"time"

	"gopractice/internal/content"
)

// Агрегаты витрины: сколько пользователей завершили курс и сколько его проходят.
// Считаются в Go поверх сырых фактов (та же completedSet-семантика, что и везде),
// кэшируются с TTL — /api/courses дёргается на каждую смену роута в SPA.

const aggTTL = 2 * time.Minute

type courseAgg struct {
	Completed  int // закрыли все обязательные уровни
	InProgress int // есть хоть один факт, но курс не завершён
}

type aggCacheEntry struct {
	at  time.Time
	agg courseAgg
}

func (a *API) courseAggregates(c *content.Course) (courseAgg, error) {
	a.aggMu.Lock()
	if e, ok := a.aggCache[c.ID]; ok && time.Since(e.at) < aggTTL {
		a.aggMu.Unlock()
		return e.agg, nil
	}
	a.aggMu.Unlock()

	byUser, err := a.store.CourseFactsAll(c.ID)
	if err != nil {
		return courseAgg{}, err
	}
	var agg courseAgg
	for _, facts := range byUser {
		switch {
		case courseCompletedFor(c, facts):
			agg.Completed++
		case len(facts.Levels) > 0 || len(facts.Tasks) > 0:
			agg.InProgress++
		}
	}

	a.aggMu.Lock()
	a.aggCache[c.ID] = aggCacheEntry{at: time.Now(), agg: agg}
	a.aggMu.Unlock()
	return agg, nil
}

// invalidateAggregates сбрасывает кэш; вызывается после reload контента —
// смена состава уровней меняет и определение завершённости.
func (a *API) invalidateAggregates() {
	a.aggMu.Lock()
	a.aggCache = map[string]aggCacheEntry{}
	a.aggMu.Unlock()
}
