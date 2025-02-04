-- name: ListUserStats :many
SELECT * FROM user_stats;

-- name: GetStatsByUserId :one
SELECT * FROM user_stats WHERE user_id = $1 LIMIT 1;

-- name: UpsertUserStats :one
INSERT INTO user_stats (
    user_id,
    rating_avg,
    rating_count
) VALUES (
    $1, $2, $3
)
ON CONFLICT (user_id) DO UPDATE
SET 
    rating_avg = EXCLUDED.rating_avg,
    rating_count = EXCLUDED.rating_count,
    updated_at = NOW()
RETURNING *;