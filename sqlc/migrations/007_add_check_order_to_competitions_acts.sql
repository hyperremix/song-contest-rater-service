ALTER TABLE competitions_acts
ADD CONSTRAINT check_order_minimum
CHECK (
    "order" >= 1
);

---- create above / drop below ----

ALTER TABLE competitions_acts
DROP CONSTRAINT check_order_minimum;
