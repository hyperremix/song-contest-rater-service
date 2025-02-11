ALTER TABLE competitions_acts DROP CONSTRAINT competitions_acts_competition_id_order_key;

---- create above / drop below ----

ALTER TABLE competitions_acts ADD CONSTRAINT competitions_acts_competition_id_order_key UNIQUE (competition_id, "order");
