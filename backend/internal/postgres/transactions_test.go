package postgres_test

import (
	"errors"
	"math"
	"smart_budget/internal/db/apidb"
	"smart_budget/internal/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestTransactionsStorage(t *testing.T) {
	pool := testutil.Database(t)
	ctx := t.Context()
	_, err := pool.Exec(ctx, `INSERT INTO app.users(id,email,password_hash) VALUES
 ('a','a@example.test','unused'),('b','b@example.test','unused');
 INSERT INTO app.ledger_periods(user_id,month,timezone) VALUES
 ('a','2026-11-01','Asia/Novosibirsk'),('a','2026-10-01','Asia/Novosibirsk'),('b','2026-11-01','UTC')`)
	if err != nil {
		t.Fatal(err)
	}

	q := apidb.New(pool)

	for _, user := range []string{"a", "b"} {
		if _, err := q.CreateBankConnection(ctx, apidb.CreateBankConnectionParams{ID: user, UserID: user, Provider: "mock", ExternalID: "account"}); err != nil {
			t.Fatal(err)
		}
	}

	month := pgtype.Date{Time: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	p := apidb.CreateTransactionParams{ID: "first", UserID: "a", BankConnectionID: "a", ExternalID: "external", Month: month, OccurredAt: time.Date(2026, 10, 31, 18, 0, 0, 0, time.UTC), Kind: "expense", AmountKopecks: math.MaxInt64, Currency: "RUB"}
	if _, err := q.CreateTransaction(ctx, p); err != nil {
		t.Fatal(err)
	}

	if _, err := q.CreateTransaction(ctx, p); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("duplicate: %v", err)
	}

	if _, err := q.GetTransaction(ctx, apidb.GetTransactionParams{UserID: "b", ID: p.ID}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("foreign owner: %v", err)
	}

	p.ID = "second"
	p.ExternalID = "second"
	if _, err := q.CreateTransaction(ctx, p); err != nil {
		t.Fatal(err)
	}

	for _, kind := range []string{"income", "transfer"} {
		p.ID = kind
		p.ExternalID = kind
		p.Kind = kind
		if _, err := q.CreateTransaction(ctx, p); err != nil {
			t.Fatal(err)
		}
	}

	total, err := q.GetMonthlyExpense(ctx, apidb.GetMonthlyExpenseParams{UserID: "a", Month: month})
	if err != nil || total != "18446744073709551614" {
		t.Fatalf("total=%s err=%v", total, err)
	}

	p.ID = "invalid"
	p.ExternalID = "invalid"
	p.Kind = "expense"

	for _, tc := range []struct {
		name   string
		change func(*apidb.CreateTransactionParams)
	}{
		{"foreign connection", func(p *apidb.CreateTransactionParams) { p.BankConnectionID = "b" }},
		{"zero", func(p *apidb.CreateTransactionParams) { p.AmountKopecks = 0 }},
		{"negative", func(p *apidb.CreateTransactionParams) { p.AmountKopecks = -1 }},
		{"currency", func(p *apidb.CreateTransactionParams) { p.Currency = "USD" }},
		{"kind", func(p *apidb.CreateTransactionParams) { p.Kind = "refund" }},
		{"mcc", func(p *apidb.CreateTransactionParams) { p.Mcc = pgtype.Text{String: "123", Valid: true} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			invalid := p
			tc.change(&invalid)
			if _, err := q.CreateTransaction(ctx, invalid); err == nil {
				t.Fatal("invalid operation accepted")
			}
		})
	}

	p.Month = pgtype.Date{Time: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	if _, err := q.CreateTransaction(ctx, p); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("wrong local month: %v", err)
	}
}
