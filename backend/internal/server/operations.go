package server

import (
	"net/http"
	"smart_budget/internal/importer"
)

// @Summary Операции
// @Description Возвращает последние 100 операций пользователя, сначала новые. Сумма передаётся в копейках строкой.
// @Tags Операции
// @Produce json
// @Security accessCookie
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Success 200 {array} importer.Operation "Список операций"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/operations [get]
func (s *Server) listOperations(w http.ResponseWriter, r *http.Request) {
	user, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	var operations []importer.Operation

	operations, err = s.Imports.ListOperations(r.Context(), user)
	if err != nil {
		respondError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, operations)
}
