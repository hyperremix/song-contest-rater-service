ALTER TABLE ratings ADD COLUMN total INT GENERATED ALWAYS AS (song + singing + "show" + looks + clothes) STORED;

---- create above / drop below ----

ALTER TABLE ratings DROP COLUMN total;
