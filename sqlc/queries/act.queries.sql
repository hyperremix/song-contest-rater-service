-- name: ListActs :many
SELECT * FROM acts;

-- name: ListActsByContestId :many
SELECT a.*, p.order FROM acts a
JOIN participations p ON a.id = p.act_id
WHERE p.contest_id = $1;

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
    id = $4 RETURNING *;

-- name: DeleteActById :one
DELETE FROM acts WHERE id = $1 RETURNING *;