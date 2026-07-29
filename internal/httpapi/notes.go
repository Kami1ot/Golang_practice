package httpapi

// Заметки: Markdown-заметка пользователя к уровню (одна на уровень).
// Хранится только сырой факт (user_id, курс, уровень, текст); имя для кабинета
// «{курс} — {модуль} — {уровень}» собирается из текущего каталога.

import (
	"net/http"
	"time"
)

type noteDTO struct {
	Body      string     `json:"body"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

func (a *API) handleNoteGet(w http.ResponseWriter, r *http.Request) {
	c, l, _, ok := a.levelFromPath(w, r)
	if !ok {
		return
	}
	n, err := a.store.Note(userFrom(r).ID, c.ID, l.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	dto := noteDTO{}
	if n != nil {
		dto.Body = n.Body
		dto.UpdatedAt = &n.UpdatedAt
	}
	writeJSON(w, http.StatusOK, dto)
}

type noteSubmission struct {
	Body string `json:"body"`
}

func (a *API) handleNotePut(w http.ResponseWriter, r *http.Request) {
	c, l, _, ok := a.levelFromPath(w, r)
	if !ok {
		return
	}
	var sub noteSubmission
	if !readJSON(w, r, &sub) {
		return
	}
	if err := a.store.SaveNote(userFrom(r).ID, c.ID, l.ID, sub.Body); err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// noteRefDTO — элемент списка заметок в кабинете. Заголовки резолвятся по
// текущему каталогу; если курс/уровень пропал из контента, показываем id.
type noteRefDTO struct {
	CourseID    string    `json:"courseId"`
	LevelID     string    `json:"levelId"`
	CourseTitle string    `json:"courseTitle"`
	Group       string    `json:"group,omitempty"`
	LevelTitle  string    `json:"levelTitle"`
	Body        string    `json:"body"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (a *API) handleNotesList(w http.ResponseWriter, r *http.Request) {
	notes, err := a.store.NotesAll(userFrom(r).ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	cat := a.cat()
	list := make([]noteRefDTO, 0, len(notes))
	for _, n := range notes {
		ref := noteRefDTO{
			CourseID: n.CourseID, LevelID: n.LevelID,
			CourseTitle: n.CourseID, LevelTitle: n.LevelID,
			Body: n.Body, UpdatedAt: n.UpdatedAt,
		}
		if c, ok := cat.Courses[n.CourseID]; ok {
			ref.CourseTitle = c.Title
			if l, ok := c.Levels[n.LevelID]; ok {
				ref.LevelTitle = l.Title
				ref.Group = l.Group
			}
		}
		list = append(list, ref)
	}
	writeJSON(w, http.StatusOK, list)
}
