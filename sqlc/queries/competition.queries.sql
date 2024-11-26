-- name: ListCompetitions :many
SELECT * FROM competitions ORDER BY start_time ASC;

-- name: GetCompetitionById :one
SELECT * FROM competitions WHERE id = $1 LIMIT 1;

-- name: InsertCompetition :one
INSERT INTO
    competitions (city, country, "description", start_time, image_url)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateCompetition :one
UPDATE
    competitions
SET
    city = $1,
    country = $2,
    "description" = $3,
    start_time = $4,
    image_url = $5,
    updated_at = NOW()
WHERE
    id = $6 RETURNING *;

-- name: DeleteCompetitionById :one
DELETE FROM competitions WHERE id = $1 RETURNING *;