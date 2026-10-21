package importer

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"smart_budget/internal/db/apidb"
	"smart_budget/internal/tbankcsv"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MaxUploadBytes  int64 = 10 << 20
	maxStoredErrors       = 100
)

var (
	ErrFileTooLarge          = errors.New("file_too_large")
	ErrInvalidIdempotencyKey = errors.New("invalid_idempotency_key")
	ErrImportNotFound        = errors.New("import_not_found")
	ErrImportInProgress      = errors.New("import_in_progress")
)

type Input struct {
	Data           []byte
	FileName       string
	IdempotencyKey string
}

type Import struct {
	ID              string              `json:"id"`
	FileName        string              `json:"fileName"`
	FileHash        string              `json:"fileHash"`
	Status          string              `json:"status" enums:"processing,completed,failed"`
	TotalRows       int                 `json:"totalRows"`
	ImportedRows    int                 `json:"importedRows"`
	SkippedRows     int                 `json:"skippedRows"`
	DuplicateRows   int                 `json:"duplicateRows"`
	ConflictRows    int                 `json:"conflictRows"`
	FailedRows      int                 `json:"failedRows"`
	Errors          []tbankcsv.RowError `json:"errors"`
	ErrorsTruncated bool                `json:"errorsTruncated"`
	CreatedAt       time.Time           `json:"createdAt" format:"date-time"`
	FinishedAt      *time.Time          `json:"finishedAt,omitempty" format:"date-time" validate:"optional"`
}

type Operation struct {
	ID         string    `json:"id"`
	ExternalID string    `json:"externalId"`
	OccurredAt time.Time `json:"occurredAt" format:"date-time"`
	Month      string    `json:"month" example:"2025-08"`
	Amount     string    `json:"amount" description:"Сумма в копейках"`
	Currency   string    `json:"currency"`
	Merchant   string    `json:"merchant"`
	MCC        string    `json:"mcc"`
	ImportID   string    `json:"importId"`
	CreatedAt  time.Time `json:"createdAt" format:"date-time"`
}

type Service struct {
	db             *pgxpool.Pool
	maxUploadBytes int64
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db, maxUploadBytes: MaxUploadBytes}
}

func (s *Service) Import(ctx context.Context, userID string, input Input) (Import, bool, error) {
	if !validIdempotencyKey(input.IdempotencyKey) {
		return Import{}, false, ErrInvalidIdempotencyKey
	}

	if input.IdempotencyKey != "" {
		imported, found, err := s.replay(ctx, userID, input.IdempotencyKey)
		if err != nil || found {
			return imported, found, err
		}
	}

	if int64(len(input.Data)) > s.maxUploadBytes {
		return Import{}, false, ErrFileTooLarge
	}

	fileName := strings.TrimSpace(input.FileName)
	if fileName == "" {
		fileName = "operations.csv"
	}

	fileHash := sha256.Sum256(input.Data)
	parsed, err := tbankcsv.Parse(bytes.NewReader(input.Data))
	if err != nil {
		return Import{}, false, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Import{}, false, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	q := apidb.New(tx)
	importID := rand.Text()
	if err := q.CreateImport(ctx, apidb.CreateImportParams{
		ID:             importID,
		UserID:         userID,
		IdempotencyKey: pgtype.Text{String: input.IdempotencyKey, Valid: input.IdempotencyKey != ""},
		FileName:       fileName,
		FileHash:       hex.EncodeToString(fileHash[:]),
	}); err != nil {
		if input.IdempotencyKey != "" && isUniqueViolation(err) {
			_ = tx.Rollback(ctx)

			imported, found, replayErr := s.replay(ctx, userID, input.IdempotencyKey)
			if replayErr != nil || found {
				return imported, found, replayErr
			}
		}

		return Import{}, false, err
	}

	timezone, err := q.GetUserTimezone(ctx, userID)
	if err != nil {
		return Import{}, false, err
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return Import{}, false, fmt.Errorf("invalid user timezone: %w", err)
	}

	candidates, duplicates, fileConflicts := prepareRows(parsed.Rows, location)
	conflicts := len(fileConflicts)
	collected := newErrorCollector(parsed.Errors, parsed.ErrorsTruncated)

	for _, conflict := range fileConflicts {
		collected.add(conflict.line, conflict.message)
	}

	for _, month := range candidateMonths(candidates) {
		date := pgtype.Date{Time: month, Valid: true}
		if err := q.EnsureLedgerPeriod(ctx, apidb.EnsureLedgerPeriodParams{UserID: userID, Month: date, Timezone: timezone}); err != nil {
			return Import{}, false, err
		}

		if _, err := q.LockLedgerPeriod(ctx, apidb.LockLedgerPeriodParams{UserID: userID, Month: date}); err != nil {
			return Import{}, false, err
		}
	}

	existing := make(map[string]string, len(candidates))
	if len(candidates) > 0 {
		externalIDs := make([]string, 0, len(candidates))

		for _, candidate := range candidates {
			externalIDs = append(externalIDs, candidate.row.ExternalID)
		}

		rows, err := q.FindOperationsByExternalIDs(ctx, apidb.FindOperationsByExternalIDsParams{UserID: userID, ExternalIds: externalIDs})
		if err != nil {
			return Import{}, false, err
		}

		for _, row := range rows {
			existing[row.ExternalID] = row.RowHash
		}
	}

	importedRows := 0

	for _, candidate := range candidates {
		if rowHash, found := existing[candidate.row.ExternalID]; found {
			if rowHash == candidate.row.RowHash {
				duplicates++
			} else {
				conflicts++

				collected.add(candidate.row.Line, "операция с тем же external_id уже существует с другими данными")
			}

			continue
		}

		if err := q.InsertOperation(ctx, apidb.InsertOperationParams{
			ID:            rand.Text(),
			UserID:        userID,
			ImportID:      importID,
			ExternalID:    candidate.row.ExternalID,
			OccurredAt:    candidate.row.OccurredAt,
			Month:         pgtype.Date{Time: candidate.month, Valid: true},
			AmountKopecks: candidate.row.AmountKopecks,
			Currency:      candidate.row.Currency,
			Merchant:      candidate.row.Merchant,
			Mcc:           candidate.row.MCC,
			RowHash:       candidate.row.RowHash,
		}); err != nil {
			return Import{}, false, err
		}

		importedRows++
	}

	errorsJSON, err := json.Marshal(collected.errors)
	if err != nil {
		return Import{}, false, err
	}

	if err := q.CompleteImport(ctx, apidb.CompleteImportParams{
		TotalRows:       int32(parsed.TotalRows),
		ImportedRows:    int32(importedRows),
		SkippedRows:     int32(parsed.SkippedRows),
		DuplicateRows:   int32(duplicates),
		ConflictRows:    int32(conflicts),
		FailedRows:      int32(parsed.FailedRows),
		Errors:          errorsJSON,
		ErrorsTruncated: collected.truncated,
		ID:              importID,
		UserID:          userID,
	}); err != nil {
		return Import{}, false, err
	}

	audit, err := json.Marshal(map[string]any{
		"fileName":        fileName,
		"fileHash":        hex.EncodeToString(fileHash[:]),
		"totalRows":       parsed.TotalRows,
		"importedRows":    importedRows,
		"skippedRows":     parsed.SkippedRows,
		"duplicateRows":   duplicates,
		"conflictRows":    conflicts,
		"failedRows":      parsed.FailedRows,
		"errorsTruncated": collected.truncated,
	})
	if err != nil {
		return Import{}, false, err
	}

	if err := q.CreateAuditEntry(ctx, apidb.CreateAuditEntryParams{
		ID:         rand.Text(),
		UserID:     userID,
		ActorID:    userID,
		EntityType: "import",
		EntityID:   importID,
		Action:     "import.completed",
		NewValue:   audit,
	}); err != nil {
		return Import{}, false, err
	}

	row, err := q.GetImport(ctx, apidb.GetImportParams{ID: importID, UserID: userID})
	if err != nil {
		return Import{}, false, err
	}

	imported, err := convertImport(row)
	if err != nil {
		return Import{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Import{}, false, err
	}

	return imported, false, nil
}

func (s *Service) GetImport(ctx context.Context, userID, id string) (Import, error) {
	row, err := apidb.New(s.db).GetImport(ctx, apidb.GetImportParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Import{}, ErrImportNotFound
	}

	if err != nil {
		return Import{}, err
	}

	return convertImport(row)
}

func (s *Service) ListImports(ctx context.Context, userID string) ([]Import, error) {
	rows, err := apidb.New(s.db).ListImports(ctx, userID)
	if err != nil {
		return nil, err
	}

	imports := make([]Import, 0, len(rows))

	for _, row := range rows {
		imported, err := convertImport(row)
		if err != nil {
			return nil, err
		}

		imports = append(imports, imported)
	}

	return imports, nil
}

func (s *Service) ListOperations(ctx context.Context, userID string) ([]Operation, error) {
	rows, err := apidb.New(s.db).ListOperations(ctx, userID)
	if err != nil {
		return nil, err
	}

	operations := make([]Operation, 0, len(rows))

	for _, row := range rows {
		operations = append(operations, Operation{
			ID:         row.ID,
			ExternalID: row.ExternalID,
			OccurredAt: row.OccurredAt.UTC(),
			Month:      row.Month.Time.Format("2006-01"),
			Amount:     strconv.FormatInt(row.AmountKopecks, 10),
			Currency:   row.Currency,
			Merchant:   row.Merchant,
			MCC:        row.Mcc,
			ImportID:   row.ImportID,
			CreatedAt:  row.CreatedAt.UTC(),
		})
	}

	return operations, nil
}

func (s *Service) replay(ctx context.Context, userID, key string) (Import, bool, error) {
	id, err := apidb.New(s.db).FindImportIDByIdempotencyKey(ctx, apidb.FindImportIDByIdempotencyKeyParams{
		UserID:         userID,
		IdempotencyKey: pgtype.Text{String: key, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Import{}, false, nil
	}

	if err != nil {
		return Import{}, false, err
	}

	imported, err := s.GetImport(ctx, userID, id)
	if err != nil {
		return Import{}, false, err
	}

	if imported.Status == "processing" {
		return Import{}, true, ErrImportInProgress
	}

	return imported, true, nil
}

type candidate struct {
	row   tbankcsv.Row
	month time.Time
}

type conflict struct {
	line    int
	message string
}

func prepareRows(rows []tbankcsv.Row, location *time.Location) ([]candidate, int, []conflict) {
	candidates := make([]candidate, 0, len(rows))
	seen := make(map[string]int, len(rows))
	conflicts := make([]conflict, 0)
	duplicates := 0

	for _, row := range rows {
		if index, found := seen[row.ExternalID]; found {
			if candidates[index].row.RowHash == row.RowHash {
				duplicates++

				continue
			}

			conflicts = append(conflicts, conflict{line: row.Line, message: "external_id повторяется в файле с другими данными"})

			continue
		}

		seen[row.ExternalID] = len(candidates)
		candidates = append(candidates, candidate{row: row, month: monthStart(row.OccurredAt, location)})
	}

	return candidates, duplicates, conflicts
}

func candidateMonths(candidates []candidate) []time.Time {
	unique := make(map[time.Time]struct{}, len(candidates))

	for _, candidate := range candidates {
		unique[candidate.month] = struct{}{}
	}

	months := make([]time.Time, 0, len(unique))

	for month := range unique {
		months = append(months, month)
	}

	sort.Slice(months, func(i, j int) bool { return months[i].Before(months[j]) })

	return months
}

func monthStart(value time.Time, location *time.Location) time.Time {
	local := value.In(location)

	return time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, time.UTC)
}

type errorCollector struct {
	errors    []tbankcsv.RowError
	truncated bool
}

func newErrorCollector(initial []tbankcsv.RowError, truncated bool) *errorCollector {
	collected := make([]tbankcsv.RowError, 0, len(initial))
	collected = append(collected, initial...)

	return &errorCollector{errors: collected, truncated: truncated}
}

func (c *errorCollector) add(line int, message string) {
	if len(c.errors) >= maxStoredErrors {
		c.truncated = true

		return
	}

	c.errors = append(c.errors, tbankcsv.RowError{Line: line, Message: message})
}

func convertImport(row apidb.AppImport) (Import, error) {
	rowErrors := make([]tbankcsv.RowError, 0)
	if len(row.Errors) > 0 {
		if err := json.Unmarshal(row.Errors, &rowErrors); err != nil {
			return Import{}, fmt.Errorf("decode import errors: %w", err)
		}
	}

	finishedAt := row.FinishedAt
	if finishedAt != nil {
		value := finishedAt.UTC()
		finishedAt = &value
	}

	return Import{
		ID:              row.ID,
		FileName:        row.FileName,
		FileHash:        row.FileHash,
		Status:          row.Status,
		TotalRows:       int(row.TotalRows),
		ImportedRows:    int(row.ImportedRows),
		SkippedRows:     int(row.SkippedRows),
		DuplicateRows:   int(row.DuplicateRows),
		ConflictRows:    int(row.ConflictRows),
		FailedRows:      int(row.FailedRows),
		Errors:          rowErrors,
		ErrorsTruncated: row.ErrorsTruncated,
		CreatedAt:       row.CreatedAt.UTC(),
		FinishedAt:      finishedAt,
	}, nil
}

func validIdempotencyKey(key string) bool {
	return key == "" || (len(key) >= 16 && len(key) <= 100)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
