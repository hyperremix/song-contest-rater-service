CREATE TYPE HEAT AS ENUM (
    'HEAT_UNSPECIFIED',
    'HEAT_SEMI_FINAL',
    'HEAT_FINAL',
    'HEAT_1',
    'HEAT_2',
    'HEAT_3',
    'HEAT_4',
    'HEAT_5',
    'HEAT_FINAL_QUALIFIER'
);
ALTER TABLE competitions ADD COLUMN heat HEAT NOT NULL DEFAULT 'HEAT_UNSPECIFIED';
ALTER TABLE competitions DROP COLUMN "description";

---- create above / drop below ----

ALTER TABLE competitions DROP COLUMN heat;
ALTER TABLE competitions ADD COLUMN "description" TEXT NOT NULL;
DROP TYPE HEAT;