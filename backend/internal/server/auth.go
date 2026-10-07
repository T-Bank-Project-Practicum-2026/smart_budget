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

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	session, err := s.Auth.Refresh(r.Context(), cookieValue(r, refreshCookie))
	if err != nil {
		respondError(w, err)

		return
	}

	s.setAuthCookies(w, session)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if err := s.Auth.Logout(r.Context(), cookieValue(r, accessCookie), cookieValue(r, refreshCookie), r.Header.Get("X-Account-ID")); err != nil {
		respondError(w, err)

		return
	}

	s.cookie(w, accessCookie, "", -1, "/")
	s.cookie(w, refreshCookie, "", -1, "/api/v1/auth")
	w.WriteHeader(http.StatusNoContent)
}

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
