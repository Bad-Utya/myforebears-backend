CREATE TABLE IF NOT EXISTS visualisations
(
    id                  UUID PRIMARY KEY,
    owner_user_id       INT NOT NULL,
    tree_id             UUID NOT NULL,
    root_person_id      UUID NOT NULL,
    included_person_ids TEXT[] NOT NULL DEFAULT '{}',
    type                TEXT NOT NULL,
    status              TEXT NOT NULL,
    file_name           TEXT NOT NULL,
    mime_type           TEXT NOT NULL,
    size_bytes          BIGINT NOT NULL DEFAULT 0,
    object_key          TEXT NOT NULL UNIQUE,
    error_message       TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_visualisations_tree_id ON visualisations (tree_id);
CREATE INDEX IF NOT EXISTS idx_visualisations_status ON visualisations (status);
