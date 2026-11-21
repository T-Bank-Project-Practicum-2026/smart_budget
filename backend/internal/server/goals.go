package server

import (
	"net/http"
	"smart_budget/internal/goals"
)

// @Summary Создать цель накопления
// @Description Создаёт финансовую цель (до 1 000 000 рублей в копейках, максимум 10 целей на пользователя).
// @Tags Цели
// @Produce json
// @Accept json
// @Param body body goals.CreateInput true "Параметры цели"
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Security accessCookie
// @Success 201 {object} goals.Goal "Цель создана"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/goals [post]
func (s *Server) createGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	var input goals.CreateInput

	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, err)

		return
	}

	goal, err := s.Goals.Create(r.Context(), userID, input)
	if err != nil {
		respondError(w, err)

		return
	}

	writeJSON(w, http.StatusCreated, goal)
}

// @Summary Список целей пользователя
// @Description Возвращает все цели пользователя с расчётом рекомендуемого ежемесячного взноса.
// @Tags Цели
// @Produce json
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Security accessCookie
// @Success 200 {array} goals.Goal "Список целей"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/goals [get]
func (s *Server) listGoals(w http.ResponseWriter, r *http.Request) {
	userID, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	items, err := s.Goals.List(r.Context(), userID)
	if err != nil {
		respondError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, items)
}

// @Summary Пополнить цель
// @Description Увеличивает текущий прогресс накопления цели на указанную сумму в копейках.
// @Tags Цели
// @Produce json
// @Accept json
// @Param id path string true "ID цели"
// @Param body body goals.ProgressInput true "Сумма пополнения в копейках"
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Security accessCookie
// @Success 200 {object} goals.Goal "Обновлённая цель"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/goals/{id}/progress [patch]
func (s *Server) addGoalProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	var input goals.ProgressInput

	if err := decodeJSON(w, r, &input); err != nil {
		respondError(w, err)

		return
	}

	goal, err := s.Goals.AddProgress(r.Context(), userID, r.PathValue("id"), input)
	if err != nil {
		respondError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, goal)
}

// @Summary Удалить цель
// @Description Удаляет финансовую цель пользователя.
// @Tags Цели
// @Produce json
// @Param id path string true "ID цели"
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Security accessCookie
// @Success 204 "Выполнено, без тела ответа"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/goals/{id} [delete]
func (s *Server) deleteGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	if err := s.Goals.Delete(r.Context(), userID, r.PathValue("id")); err != nil {
		respondError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
