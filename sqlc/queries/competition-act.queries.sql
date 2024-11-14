-- name: InsertCompetitionAct :exec
INSERT INTO
    competitions_acts (competition_id, act_id)
VALUES ($1, $2);

-- name: DeleteCompetitionAct :exec
DELETE FROM competitions_acts WHERE competition_id = $1 AND act_id = $2;
