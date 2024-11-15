-- name: ListUsers :many
SELECT * FROM users;

-- name: ListUsersByCompetitionId :many
SELECT u.* FROM users u
JOIN ratings r ON u.id = r.user_id
WHERE r.competition_id = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserBySub :one
SELECT * FROM users WHERE sub = $1 LIMIT 1;

-- name: InsertUser :one
INSERT INTO
    users (sub, email, firstname, lastname, image_url)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateUser :one
UPDATE
    users
SET
    firstname = $1,
    lastname = $2,
    image_url = $3,
    updated_at = NOW()
WHERE
    id = $4 RETURNING *;

-- name: DeleteUserById :one
DELETE FROM users WHERE id = $1 RETURNING *;