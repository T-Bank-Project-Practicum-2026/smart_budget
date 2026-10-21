package server

import (
	"errors"
	"io"
	"net/http"
	"smart_budget/internal/importer"
	"strings"
)

// @Summary Импорт операций T-Bank
// @Description Принимает CSV-выписку T-Bank. Повторные операции пропускаются, конфликтующие строки сохраняются в ошибках импорта.
// @Tags Импорт
// @Accept multipart/form-data
// @Produce json
// @Security accessCookie
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Param Idempotency-Key header string false "Ключ идемпотентности" minLength(16) maxLength(100)
// @Param file formData file true "CSV-выписка T-Bank, не больше 10 МиБ"
// @Success 201 {object} importer.Import "Импорт создан"
// @Success 200 {object} importer.Import "Повторный запрос вернул существующий импорт"
// @Failure 400 {object} ErrorResponse "Некорректный файл или ключ идемпотентности"
// @Failure 413 {object} ErrorResponse "Файл больше 10 МиБ"
// @Failure 409 {object} ErrorResponse "Импорт с этим ключом ещё выполняется"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/imports/tbank [post]
func (s *Server) createTBankImport(w http.ResponseWriter, r *http.Request) {
	user, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	data, fileName, err := readTBankFile(w, r)
	if err != nil {
		respondError(w, err)

		return
	}

	imported, replayed, err := s.Imports.Import(r.Context(), user, importer.Input{
		Data:           data,
		FileName:       fileName,
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		respondError(w, err)

		return
	}

	status := http.StatusCreated
	if replayed {
		status = http.StatusOK
	}

	writeJSON(w, status, imported)
}

// @Summary История импортов
// @Description Возвращает последние 50 импортов пользователя, сначала новые.
// @Tags Импорт
// @Produce json
// @Security accessCookie
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Success 200 {array} importer.Import "Список импортов"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/imports [get]
func (s *Server) listImports(w http.ResponseWriter, r *http.Request) {
	user, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	imports, err := s.Imports.ListImports(r.Context(), user)
	if err != nil {
		respondError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, imports)
}

// @Summary Импорт по ID
// @Tags Импорт
// @Produce json
// @Security accessCookie
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Param id path string true "ID импорта"
// @Success 200 {object} importer.Import "Импорт"
// @Failure 404 {object} ErrorResponse "Импорт не найден"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/imports/{id} [get]
func (s *Server) getImport(w http.ResponseWriter, r *http.Request) {
	user, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	imported, err := s.Imports.GetImport(r.Context(), user, r.PathValue("id"))
	if err != nil {
		respondError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, imported)
}

func readTBankFile(w http.ResponseWriter, r *http.Request) ([]byte, string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, importer.MaxUploadBytes+(1<<20))

	reader, err := r.MultipartReader()
	if err != nil {
		return nil, "", errInvalidRequest
	}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			return nil, "", errInvalidRequest
		}

		if err != nil {
			return nil, "", uploadError(err)
		}

		if part.FormName() != "file" {
			_ = part.Close()

			continue
		}

		fileName := strings.TrimSpace(part.FileName())
		data, err := io.ReadAll(io.LimitReader(part, importer.MaxUploadBytes+1))
		_ = part.Close()

		if err != nil {
			return nil, "", uploadError(err)
		}

		if int64(len(data)) > importer.MaxUploadBytes {
			return nil, "", importer.ErrFileTooLarge
		}

		if fileName == "" {
			fileName = "operations.csv"
		}

		return data, fileName, nil
	}
}

func uploadError(err error) error {
	var maxBytes *http.MaxBytesError

	if errors.As(err, &maxBytes) {
		return importer.ErrFileTooLarge
	}

	return errInvalidRequest
}
