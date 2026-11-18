package postgres_test

import (
	"smart_budget/internal/db/apidb"
	"smart_budget/internal/importer"
	"smart_budget/internal/testutil"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const csvHeader = "Имя счёта;Номер карты;Дата операции;Сумма в валюте счёта;Валюта счёта;Статус;Категория по-умолчанию;MCC;Описание;Учёт в аналитике\n"

func TestOperationsFromCSV(t *testing.T) {
	pool := testutil.Database(t)
	ctx := t.Context()

	_, err := pool.Exec(ctx, `INSERT INTO app.users(id,email,password_hash,timezone) VALUES
 ('a','a@example.test','unused','Asia/Novosibirsk'),('b','b@example.test','unused','UTC')`)
	if err != nil {
		t.Fatal(err)
	}

	data := csvHeader +
		"Счёт;*1234;31.10.2026 22:00:00;-92233720368547758,08;RUB;Ок;Продукты;5411;Магазин;Да\n" +
		"Счёт;*1234;31.10.2026 22:01:00;-92233720368547758,07;RUB;Ок;Продукты;5411;Магазин;Да\n" +
		"Счёт;*1234;31.10.2026 22:02:00;-300;RUB;Ок;Переводы;;Перевод;Да\n" +
		"Счёт;*1234;31.10.2026 22:03:00;300;RUB;Ок;Продукты;;Доход;Да\n"
	service := importer.New(pool)
	input := importer.Input{Data: []byte(data), FileName: "operations.csv"}

	var group sync.WaitGroup

	var imports [2]importer.Import

	var importErrors [2]error

	for i := range imports {
		group.Go(func() {
			imports[i], _, importErrors[i] = service.Import(ctx, "a", input)
		})
	}

	group.Wait()

	for _, importErr := range importErrors {
		if importErr != nil {
			t.Fatal(importErr)
		}
	}

	if imports[0].ImportedRows+imports[1].ImportedRows != 2 ||
		imports[0].DuplicateRows+imports[1].DuplicateRows != 2 {
		t.Fatalf("concurrent imports: %+v", imports)
	}

	conflict := importer.Input{
		Data:     []byte(strings.Replace(data, ";5411;Магазин;", ";5412;Магазин;", 1)),
		FileName: input.FileName,
	}
	report, _, err := service.Import(ctx, "a", conflict)
	if err != nil || report.ConflictRows != 1 || report.DuplicateRows != 1 {
		t.Fatalf("conflicting import: %+v, %v", report, err)
	}

	month := pgtype.Date{Time: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	q := apidb.New(pool)
	total, err := q.GetMonthlyExpense(ctx, apidb.GetMonthlyExpenseParams{UserID: "a", Month: month})
	if err != nil || total != "18446744073709551615" {
		t.Fatalf("expense: %q, %v", total, err)
	}

	total, err = q.GetMonthlyExpense(ctx, apidb.GetMonthlyExpenseParams{UserID: "b", Month: month})
	if err != nil || total != "0" {
		t.Fatalf("other account expense: %q, %v", total, err)
	}

	rows, err := service.ListOperations(ctx, "a")
	if err != nil || len(rows) != 2 || rows[0].Month != "2026-11" || rows[0].Amount != "-9223372036854775807" {
		t.Fatalf("operations: %+v, %v", rows, err)
	}

	rows, err = service.ListOperations(ctx, "b")
	if err != nil || len(rows) != 0 {
		t.Fatalf("other account operations: %+v, %v", rows, err)
	}

	// Force a failure during audit insertion, after the operation has been inserted.
	_, err = pool.Exec(ctx, `ALTER TABLE app.audit_log ADD CONSTRAINT reject_new_audit CHECK (false) NOT VALID`)
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := service.Import(ctx, "b", input); err == nil {
		t.Fatal("import succeeded despite audit failure")
	}

	var importsCount, operationsCount int

	err = pool.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM app.imports WHERE user_id='b'),
 (SELECT count(*) FROM app.operations WHERE user_id='b')`).Scan(&importsCount, &operationsCount)
	if err != nil || importsCount != 0 || operationsCount != 0 {
		t.Fatalf("rollback: imports=%d operations=%d err=%v", importsCount, operationsCount, err)
	}
}
