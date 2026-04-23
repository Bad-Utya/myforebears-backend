DROP INDEX IF EXISTS idx_persons_avatar_photo_id;

ALTER TABLE persons
    DROP COLUMN IF EXISTS avatar_photo_id;
