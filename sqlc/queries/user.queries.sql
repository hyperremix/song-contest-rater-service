-- name: ListUsers :many
SELECT * FROM users;

-- name: ListUsersByCompetitionId :many
SELECT u.* FROM users u
JOIN ratings r ON u.id = r.user_id
WHERE r.competition_id = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: InsertUser :one
INSERT INTO
    users (email, firstname, lastname, image_url)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateUser :one
UPDATE
    users
SET
    email = $1,
    firstname = $2,
    lastname = $3,
    image_url = $4,
    updated_at = NOW()
WHERE
    id = $5 RETURNING *;

-- name: DeleteUserById :one
DELETE FROM users WHERE id = $1 RETURNING *;