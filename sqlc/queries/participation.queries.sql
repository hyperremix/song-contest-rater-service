-- name: ListParticipations :many
SELECT * FROM participations;

-- name: InsertParticipation :one
INSERT INTO
    participations (contest_id, act_id, "order")
VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteParticipation :one
DELETE FROM participations WHERE contest_id = $1 AND act_id = $2 RETURNING *;
