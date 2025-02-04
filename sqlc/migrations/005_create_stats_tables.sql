CREATE TABLE user_stats (
    user_id UUID PRIMARY KEY,
    rating_avg DECIMAL(3,2),
    rating_count INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE global_stats (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE,
    rating_avg DECIMAL(3,2),
    rating_count INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT only_one_row CHECK (id)
);

---- create above / drop below ----

DROP TABLE IF EXISTS user_stats;
DROP TABLE IF EXISTS global_stats; 