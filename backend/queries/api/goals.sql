-- name: CreateGoal :one
INSERT INTO app.goals (id, user_id, name, target_kopecks, deadline)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserGoals :many
SELECT * FROM app.goals
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateGoalProgress :one
UPDATE app.goals
SET current_kopecks = current_kopecks + $1, updated_at = now()
WHERE id = $2 AND user_id = $3
RETURNING *;

-- name: CountUserGoals :one
SELECT COUNT(*) FROM app.goals
WHERE user_id = $1;

-- name: DeleteGoal :exec
DELETE FROM app.goals
WHERE id = $1 AND user_id = $2;
