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
    rating_avg = EXCLUDED.rating_avg,
    rating_count = EXCLUDED.rating_count,
    updated_at = NOW()
RETURNING *;
