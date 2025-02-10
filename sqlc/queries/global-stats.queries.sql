-- name: GetGlobalStats :one
SELECT * FROM global_stats WHERE id = TRUE LIMIT 1;

-- name: UpsertGlobalStats :one
INSERT INTO global_stats (
    id,
    rating_avg,
    rating_count
) VALUES (
    TRUE, $1, $2
)
ON CONFLICT (id) DO UPDATE
SET 
    rating_avg = $2,
    rating_count = $3,
    updated_at = NOW()
RETURNING *;
