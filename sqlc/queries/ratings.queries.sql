-- name: ListRatings :many
SELECT * FROM ratings;

-- name: ListRatingsByCompetitionId :many
SELECT * FROM ratings WHERE competition_id = $1;

-- name: ListRatingsByUserId :many
SELECT * FROM ratings WHERE user_id = $1;

-- name: ListRatingsByActId :many
SELECT * FROM ratings WHERE act_id = $1;

-- name: ListRatingsByCompetitionAndAcId :many
SELECT * FROM ratings WHERE competition_id = $1 AND act_id = $2;

-- name: GetRatingById :one
SELECT * FROM ratings WHERE id = $1 LIMIT 1;

-- name: InsertRating :one
INSERT INTO
    ratings (song, singing, "show", looks, clothes, user_id, competition_id, act_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: UpdateRating :one
UPDATE
    ratings
SET
    song = $1,
    singing = $2,
    "show" = $3,
    looks = $4,
    clothes = $5,
    updated_at = NOW()
WHERE
    id = $6 RETURNING *;

-- name: DeleteRatingById :one
DELETE FROM ratings WHERE id = $1 RETURNING *;
