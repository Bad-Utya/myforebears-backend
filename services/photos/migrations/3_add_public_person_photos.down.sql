DROP INDEX IF EXISTS uq_public_person_avatar;
DROP INDEX IF EXISTS idx_photos_public_event_id;
DROP INDEX IF EXISTS idx_photos_public_person_id;
ALTER TABLE photos DROP CONSTRAINT IF EXISTS photos_person_avatar_owner_check;
ALTER TABLE photos ADD CONSTRAINT photos_person_avatar_owner_check
    CHECK (NOT is_person_avatar OR person_id IS NOT NULL);
ALTER TABLE photos
    DROP COLUMN IF EXISTS public_event_id,
    DROP COLUMN IF EXISTS public_person_id;
