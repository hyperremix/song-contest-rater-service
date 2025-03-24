-- name: ListCompetitionActs :many
SELECT * FROM competitions_acts;

-- name: InsertCompetitionAct :exec
INSERT INTO
    competitions_acts (competition_id, act_id, "order")
VALUES ($1, $2, $3);

-- name: DeleteCompetitionAct :exec
DELETE FROM competitions_acts WHERE competition_id = $1 AND act_id = $2;
