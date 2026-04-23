CREATE TABLE IF NOT EXISTS photos
(
    id               UUID PRIMARY KEY,
    owner_user_id    INT NOT NULL,
    tree_id          UUID,
    person_id        UUID,
    event_id         UUID,
    is_user_avatar   BOOLEAN NOT NULL DEFAULT FALSE,
    is_person_avatar BOOLEAN NOT NULL DEFAULT FALSE,
    file_name        TEXT NOT NULL,
    mime_type        TEXT NOT NULL,
    size_bytes       BIGINT NOT NULL,
    object_key       TEXT NOT NULL UNIQUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK ((person_id IS NULL) OR (event_id IS NULL)),
    CHECK ((is_user_avatar AND person_id IS NULL AND event_id IS NULL) OR (NOT is_user_avatar)),
    CHECK ((is_person_avatar AND person_id IS NOT NULL) OR (NOT is_person_avatar))
);

CREATE INDEX IF NOT EXISTS idx_photos_owner_user_id ON photos (owner_user_id);
CREATE INDEX IF NOT EXISTS idx_photos_person_id ON photos (person_id);
CREATE INDEX IF NOT EXISTS idx_photos_event_id ON photos (event_id);
CREATE INDEX IF NOT EXISTS idx_photos_tree_id ON photos (tree_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_avatar_per_user
    ON photos (owner_user_id)
    WHERE is_user_avatar = TRUE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_person_avatar_per_person
    ON photos (person_id)
    WHERE is_person_avatar = TRUE;
