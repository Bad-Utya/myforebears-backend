ALTER TABLE trees
    ADD COLUMN IF NOT EXISTS root_person_id UUID;

CREATE INDEX IF NOT EXISTS idx_trees_root_person_id ON trees (root_person_id);
