package goals

import (
	"context"
	"crypto/rand"
	"errors"
	"math"
	"smart_budget/internal/db/apidb"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MaxUserGoals     = 10
	MaxTargetKopecks = 100_000_000
)

var (
	ErrInvalidGoalName   = errors.New("invalid_goal_name")
	ErrInvalidGoalAmount = errors.New("invalid_goal_amount")
	ErrDeadlineInPast    = errors.New("deadline_in_past")
	ErrMaxGoalsReached   = errors.New("max_goals_reached")
	ErrGoalNotFound      = errors.New("goal_not_found")
	ErrGoalLimitExceeded = errors.New("goal_limit_exceeded")
)

type CreateInput struct {
	Name          string     `json:"name"`
	TargetKopecks int64      `json:"targetKopecks"`
	Deadline      *time.Time `json:"deadline,omitempty" validate:"optional"`
}

type ProgressInput struct {
	AmountKopecks int64 `json:"amountKopecks"`
}

type Goal struct {
	ID                        string     `json:"id"`
	Name                      string     `json:"name"`
	TargetKopecks             int64      `json:"targetKopecks"`
	CurrentKopecks            int64      `json:"currentKopecks"`
	Deadline                  *time.Time `json:"deadline,omitempty" validate:"optional"`
	RecommendedMonthlyKopecks int64      `json:"recommendedMonthlyKopecks"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
}

type Service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Goal, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Goal{}, ErrInvalidGoalName
	}

	if input.TargetKopecks <= 0 || input.TargetKopecks > MaxTargetKopecks {
		return Goal{}, ErrInvalidGoalAmount
	}

	now := time.Now()
	if input.Deadline != nil && input.Deadline.Before(now) {
		return Goal{}, ErrDeadlineInPast
	}

	q := apidb.New(s.db)

	count, err := q.CountUserGoals(ctx, userID)
	if err != nil {
		return Goal{}, err
	}

	if count >= MaxUserGoals {
		return Goal{}, ErrMaxGoalsReached
	}

	row, err := q.CreateGoal(ctx, apidb.CreateGoalParams{
		ID:            rand.Text(),
		UserID:        userID,
		Name:          name,
		TargetKopecks: input.TargetKopecks,
		Deadline:      input.Deadline,
	})
	if err != nil {
		return Goal{}, err
	}

	return convertGoal(row, now), nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Goal, error) {
	rows, err := apidb.New(s.db).GetUserGoals(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := make([]Goal, 0, len(rows))

	for _, row := range rows {
		result = append(result, convertGoal(row, now))
	}

	return result, nil
}

func (s *Service) AddProgress(ctx context.Context, userID, goalID string, input ProgressInput) (Goal, error) {
	if input.AmountKopecks <= 0 || input.AmountKopecks > MaxTargetKopecks {
		return Goal{}, ErrInvalidGoalAmount
	}

	row, err := apidb.New(s.db).UpdateGoalProgress(ctx, apidb.UpdateGoalProgressParams{
		CurrentKopecks: input.AmountKopecks,
		ID:             goalID,
		UserID:         userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Goal{}, ErrGoalNotFound
	}

	if err != nil {
		if strings.Contains(err.Error(), "23514") {
			return Goal{}, ErrGoalLimitExceeded
		}

		return Goal{}, err
	}

	return convertGoal(row, time.Now()), nil
}

func (s *Service) Delete(ctx context.Context, userID, goalID string) error {
	return apidb.New(s.db).DeleteGoal(ctx, apidb.DeleteGoalParams{
		ID:     goalID,
		UserID: userID,
	})
}

func convertGoal(row apidb.AppGoal, now time.Time) Goal {
	g := Goal{
		ID:             row.ID,
		Name:           row.Name,
		TargetKopecks:  row.TargetKopecks,
		CurrentKopecks: row.CurrentKopecks,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}

	if row.Deadline != nil {
		t := *row.Deadline
		g.Deadline = &t

		remaining := row.TargetKopecks - row.CurrentKopecks
		if remaining > 0 && t.After(now) {
			months := math.Ceil(t.Sub(now).Hours() / (24 * 30))
			if months < 1 {
				months = 1
			}

			g.RecommendedMonthlyKopecks = int64(math.Ceil(float64(remaining) / months))
		}
	}

	return g
}
