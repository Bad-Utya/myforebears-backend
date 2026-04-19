CREATE TABLE IF NOT EXISTS trees
(
    id         UUID PRIMARY KEY,
    creator_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trees_creator_id ON trees (creator_id);
