ALTER TABLE photos
    DROP CONSTRAINT IF EXISTS chk_photos_tree_avatar;

DROP INDEX IF EXISTS uq_tree_avatar_per_tree;

ALTER TABLE photos
    DROP COLUMN IF EXISTS is_tree_avatar;
