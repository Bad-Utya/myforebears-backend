CREATE TYPE IF NOT EXISTS gender_enum AS ENUM ('MALE', 'FEMALE');

CREATE TABLE IF NOT EXISTS persons
(
    id         UUID PRIMARY KEY,
    tree_id    UUID NOT NULL,
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL,
    patronymic TEXT,
    gender     gender_enum NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_persons_tree_id ON persons (tree_id);
CREATE INDEX IF NOT EXISTS idx_persons_tree_id_id ON persons (tree_id, id);
