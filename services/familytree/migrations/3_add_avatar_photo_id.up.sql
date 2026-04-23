ALTER TABLE persons
    ADD COLUMN IF NOT EXISTS avatar_photo_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_persons_avatar_photo_id ON persons (avatar_photo_id);
