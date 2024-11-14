CREATE TABLE competitions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    city TEXT NOT NULL,
    country TEXT NOT NULL,
    "description" TEXT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE acts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    artist_name TEXT NOT NULL,
    song_name TEXT NOT NULL,
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE competitions_acts (
    competition_id uuid NOT NULL,
    act_id uuid NOT NULL,
    PRIMARY KEY (competition_id, act_id),
    FOREIGN KEY (competition_id) REFERENCES competitions(id),
    FOREIGN KEY (act_id) REFERENCES acts(id)
);

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    firstname TEXT NOT NULL,
    lastname TEXT NOT NULL,
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ratings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    song INT CHECK (song BETWEEN 1 AND 15),
    singing INT CHECK (singing BETWEEN 1 AND 15),
    "show" INT CHECK ("show" BETWEEN 1 AND 15),
    looks INT CHECK (looks BETWEEN 1 AND 15),
    clothes INT CHECK (clothes BETWEEN 1 AND 15),
    user_id uuid NOT NULL,
    act_id uuid NOT NULL,
    competition_id uuid NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (act_id) REFERENCES acts(id),
    FOREIGN KEY (competition_id) REFERENCES competitions(id),
    UNIQUE (user_id, act_id, competition_id)
);
