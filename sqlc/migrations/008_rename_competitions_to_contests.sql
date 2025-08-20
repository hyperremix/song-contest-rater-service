ALTER TABLE competitions RENAME TO contests;

ALTER TABLE competitions_acts RENAME TO participations;

ALTER TABLE participations RENAME COLUMN competition_id TO contest_id;

ALTER TABLE ratings RENAME COLUMN competition_id TO contest_id;


---- create above / drop below ----

ALTER TABLE participations RENAME COLUMN contest_id TO competition_id;

ALTER TABLE ratings RENAME COLUMN contest_id TO competition_id;

ALTER TABLE contests RENAME TO competitions;

ALTER TABLE participations RENAME TO competitions_acts;
