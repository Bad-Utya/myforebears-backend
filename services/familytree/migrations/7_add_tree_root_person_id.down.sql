DROP INDEX IF EXISTS idx_trees_root_person_id;

ALTER TABLE trees
    DROP COLUMN IF EXISTS root_person_id;
