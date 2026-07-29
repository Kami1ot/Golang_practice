package httpapi

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"gopractice/internal/auth"
	"gopractice/internal/store"
)

const (
	sessionCookie   = "gopractice_session"
	sessionLifetime = 30 * 24 * time.Hour
	// Скользящее продление: трогаем БД, только если сессия «постарела» на сутки.
	sessionRenewAfter = 24 * time.Hour
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

type ctxUserKey struct{}

// userFrom достаёт пользователя, положенного requireUser. nil вне middleware.
func userFrom(r *http.Request) *store.User {
	u, _ := r.Context().Value(ctxUserKey{}).(*store.User)
	return u
}

// sessionUser резолвит cookie → пользователь; продлевает сессию по необходимости.
func (a *API) sessionUser(r *http.Request) *store.User {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return nil
	}
	u, expires, err := a.store.UserBySession(c.Value)
	if err != nil {
		log.Printf("сессия: %v", err)
		return nil
	}
	if u == nil {
		return nil
	}
	if time.Until(expires) < sessionLifetime-sessionRenewAfter {
		if err := a.store.TouchSession(c.Value, time.Now().Add(sessionLifetime)); err != nil {
			log.Printf("продление сессии: %v", err)
		}
	}
	return u
}

func (a *API) requireUser(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := a.sessionUser(r)
		if u == nil {
			writeErr(w, http.StatusUnauthorized, "требуется вход")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), ctxUserKey{}, u)))
	})
}

func (a *API) requireAdmin(next http.HandlerFunc) http.Handler {
	return a.requireUser(func(w http.ResponseWriter, r *http.Request) {
		if !userFrom(r).IsAdmin() {
			writeErr(w, http.StatusForbidden, "нужны права администратора")
			return
		}
		next(w, r)
	})
}

// checkLocalOrigin — анти-CSRF для multipart-загрузок (JSON-запросы прикрыты
// SameSite=Lax + Content-Type). Пустой Origin (curl) допустим.
func checkLocalOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "127.0.0.1" || host == "localhost" || u.Host == r.Host
}

type userDTO struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	ChatEnabled bool   `json:"chatEnabled"`
}

func (a *API) userDTO(u *store.User) userDTO {
	return userDTO{ID: u.ID, Username: u.Username, Role: u.Role, ChatEnabled: a.chat != nil}
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if !readJSON(w, r, &c) {
		return
	}
	if !usernameRe.MatchString(c.Username) {
		writeErr(w, http.StatusBadRequest, "имя: 3–32 символа, латиница/цифры/дефис/подчёркивание")
		return
	}
	if err := auth.ValidatePassword(c.Password); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := auth.HashPassword(c.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хеширование пароля: "+err.Error())
		return
	}
	u, err := a.store.CreateUser(c.Username, hash)
	if err == store.ErrUsernameTaken {
		writeErr(w, http.StatusConflict, "это имя уже занято")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	if !a.issueSession(w, u) {
		return
	}
	log.Printf("новый пользователь: %s (роль %s)", u.Username, u.Role)
	writeJSON(w, http.StatusCreated, a.userDTO(u))
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if !readJSON(w, r, &c) {
		return
	}
	u, err := a.store.UserByName(c.Username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return
	}
	if u == nil || !auth.CheckPassword(u.PasswordHash, c.Password) {
		time.Sleep(300 * time.Millisecond) // притормаживаем перебор
		writeErr(w, http.StatusUnauthorized, "неверный логин или пароль")
		return
	}
	if err := a.store.DeleteExpiredSessions(); err != nil {
		log.Printf("чистка сессий: %v", err)
	}
	if !a.issueSession(w, u) {
		return
	}
	writeJSON(w, http.StatusOK, a.userDTO(u))
}

func (a *API) issueSession(w http.ResponseWriter, u *store.User) bool {
	token, err := auth.NewToken()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "генерация токена: "+err.Error())
		return false
	}
	if err := a.store.CreateSession(token, u.ID, time.Now().Add(sessionLifetime)); err != nil {
		writeErr(w, http.StatusInternalServerError, "хранилище: "+err.Error())
		return false
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionLifetime.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return true
}

func (a *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		if err := a.store.DeleteSession(c.Value); err != nil {
			log.Printf("удаление сессии: %v", err)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	u := a.sessionUser(r)
	if u == nil {
		writeErr(w, http.StatusUnauthorized, "требуется вход")
		return
	}
	writeJSON(w, http.StatusOK, a.userDTO(u))
}
