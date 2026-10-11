package server

import (
	"crypto/subtle"
	"net/http"
	"smart_budget/internal/auth"
	"time"
)

const (
	accessCookie  = "budget_access"
	refreshCookie = "budget_refresh"
)

// @Summary Регистрация
// @Tags Авторизация
// @Produce json
// @Accept json
// @Param body body auth.RegisterInput true "Данные пользователя"
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Success 201 {object} auth.User "Пользователь зарегистрирован"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Header 201 {string} Set-Cookie "Cookie budget_access и budget_refresh с флагом HttpOnly"
// @Router /api/v1/auth/register [post]
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var input auth.RegisterInput

	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, err)

		return
	}

	result, err := s.Auth.Register(r.Context(), input)
	if err != nil {
		respondError(w, err)

		return
	}

	s.setAuthCookies(w, result.Session)
	writeJSON(w, http.StatusCreated, result.User)
}

// @Summary Вход
// @Tags Авторизация
// @Produce json
// @Accept json
// @Param body body auth.LoginInput true "Данные пользователя"
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Success 200 {object} auth.User "Профиль пользователя"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Header 200 {string} Set-Cookie "Cookie budget_access и budget_refresh с флагом HttpOnly"
// @Router /api/v1/auth/login [post]
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var input auth.LoginInput

	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, err)

		return
	}

	result, err := s.Auth.Login(r.Context(), input)
	if err != nil {
		respondError(w, err)

		return
	}

	s.setAuthCookies(w, result.Session)
	writeJSON(w, http.StatusOK, result.User)
}

// @Summary Обновление сессии
// @Description Обновляет обе cookie по refresh-сессии.
// @Tags Авторизация
// @Produce json
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Security refreshCookie
// @Success 204 "Выполнено, без тела ответа"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Header 204 {string} Set-Cookie "Cookie budget_access и budget_refresh с флагом HttpOnly"
// @Router /api/v1/auth/refresh [post]
func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	session, err := s.Auth.Refresh(r.Context(), cookieValue(r, refreshCookie))
	if err != nil {
		respondError(w, err)

		return
	}

	s.setAuthCookies(w, session)
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Выход
// @Description Завершает сессию по любой из двух cookie и удаляет обе. Без cookie тоже возвращает 204.
// @Tags Авторизация
// @Produce json
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Success 204 "Выполнено, без тела ответа"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/auth/logout [post]
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if err := s.Auth.Logout(r.Context(), cookieValue(r, accessCookie), cookieValue(r, refreshCookie), r.Header.Get("X-Account-ID")); err != nil {
		respondError(w, err)

		return
	}

	s.cookie(w, accessCookie, "", -1, "/")
	s.cookie(w, refreshCookie, "", -1, "/api/v1/auth")
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Текущий пользователь
// @Tags Авторизация
// @Produce json
// @Security accessCookie
// @Success 200 {object} auth.User "Профиль пользователя"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/auth/me [get]
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user, err := s.Auth.Me(r.Context(), cookieValue(r, accessCookie))
	if err != nil {
		respondError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (s *Server) user(r *http.Request) (string, error) {
	identity, err := s.identity(r)

	return identity.UserID, err
}

func (s *Server) identity(r *http.Request) (auth.ActiveSession, error) {
	identity, err := s.Auth.Identity(r.Context(), cookieValue(r, accessCookie))
	if err != nil {
		return identity, err
	}

	if expected := r.Header.Get("X-Account-ID"); expected != "" && expected != identity.UserID {
		return auth.ActiveSession{}, auth.ErrAccountChanged
	}

	return identity, nil
}

func (s *Server) sessionStatus(w http.ResponseWriter, r *http.Request) {
	if s.Token == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+s.Token)) != 1 {
		writeError(w, 403, "forbidden")

		return
	}

	_, err := s.Auth.ActiveSession(r.Context(), r.PathValue("id"), r.URL.Query().Get("userId"))
	if err != nil {
		respondError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func cookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func (s *Server) cookie(w http.ResponseWriter, name, value string, maxAge int, path string) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: path, HttpOnly: true, Secure: s.SecureCookies, SameSite: http.SameSiteStrictMode, MaxAge: maxAge})
}

func (s *Server) setAuthCookies(w http.ResponseWriter, session auth.Session) {
	s.cookie(w, accessCookie, session.Access, int(auth.AccessTTL.Seconds()), "/")
	s.cookie(w, refreshCookie, session.Refresh, max(1, int(time.Until(session.ExpiresAt).Seconds())), "/api/v1/auth")
}
