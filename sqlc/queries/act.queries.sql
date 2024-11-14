-- name: ListActs :many
SELECT * FROM acts;

-- name: ListActsByCompetitionId :many
SELECT a.* FROM acts a
JOIN competitions_acts ca ON a.id = ca.act_id
WHERE ca.competition_id = $1;

-- name: GetActById :one
SELECT * FROM acts WHERE id = $1 LIMIT 1;

-- name: InsertAct :one
INSERT INTO
    acts (artist_name, song_name, image_url)
VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateAct :one
UPDATE
    acts
SET
    artist_name = $1,
    song_name = $2,
    image_url = $3,
    updated_at = NOW()
WHERE
    id = $4 AND deleted_at IS NULL RETURNING *;

-- name: DeleteActById :one
DELETE FROM acts WHERE id = $1 RETURNING *;