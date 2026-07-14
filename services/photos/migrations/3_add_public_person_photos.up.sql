ALTER TABLE photos
    ADD COLUMN IF NOT EXISTS public_person_id UUID,
    ADD COLUMN IF NOT EXISTS public_event_id UUID;

DO $$
DECLARE constraint_name TEXT;
BEGIN
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'photos'::regclass
      AND pg_get_constraintdef(oid) ILIKE '%is_person_avatar%person_id IS NOT NULL%'
    LIMIT 1;
    IF constraint_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE photos DROP CONSTRAINT %I', constraint_name);
    END IF;
END $$;

ALTER TABLE photos ADD CONSTRAINT photos_person_avatar_owner_check
    CHECK (NOT is_person_avatar OR person_id IS NOT NULL OR public_person_id IS NOT NULL);

CREATE INDEX IF NOT EXISTS idx_photos_public_person_id ON photos (public_person_id);
CREATE INDEX IF NOT EXISTS idx_photos_public_event_id ON photos (public_event_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_public_person_avatar
    ON photos (public_person_id)
    WHERE is_person_avatar = TRUE AND public_person_id IS NOT NULL;
