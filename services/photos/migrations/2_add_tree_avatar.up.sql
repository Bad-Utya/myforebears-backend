ALTER TABLE photos
    ADD COLUMN IF NOT EXISTS is_tree_avatar BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_tree_avatar_per_tree
    ON photos (tree_id)
    WHERE is_tree_avatar = TRUE;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_photos_tree_avatar'
    ) THEN
        ALTER TABLE photos
            ADD CONSTRAINT chk_photos_tree_avatar
            CHECK (
                (is_tree_avatar AND tree_id IS NOT NULL AND person_id IS NULL AND event_id IS NULL)
                OR
                (NOT is_tree_avatar)
            );
    END IF;
END $$;
