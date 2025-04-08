-- name: ListContests :many
SELECT * FROM contests ORDER BY start_time ASC;

-- name: GetContestById :one
SELECT * FROM contests WHERE id = $1 LIMIT 1;

-- name: InsertContest :one
INSERT INTO
    contests (city, country, heat, start_time, image_url)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateContest :one
UPDATE
    contests
SET
    city = $1,
    country = $2,
    heat = $3,
    start_time = $4,
    image_url = $5,
    updated_at = NOW()
WHERE
    id = $6 RETURNING *;

-- name: DeleteContestById :one
DELETE FROM contests WHERE id = $1 RETURNING *;