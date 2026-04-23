ALTER TABLE trees
    DROP COLUMN IF EXISTS is_public_on_main_page,
    DROP COLUMN IF EXISTS is_view_restricted;
