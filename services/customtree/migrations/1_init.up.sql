CREATE TABLE custom_trees (
 id UUID PRIMARY KEY, creator_id INT NOT NULL, name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
 relation_down TEXT NOT NULL, relation_up TEXT NOT NULL, root_entity_id UUID,
 is_view_restricted BOOLEAN NOT NULL DEFAULT TRUE, is_public_on_main_page BOOLEAN NOT NULL DEFAULT FALSE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_custom_trees_creator ON custom_trees(creator_id);

CREATE TABLE custom_entities (
 id UUID PRIMARY KEY, tree_id UUID NOT NULL REFERENCES custom_trees(id) ON DELETE CASCADE,
 name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', avatar_photo_id UUID,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_custom_entities_tree ON custom_entities(tree_id);

CREATE TABLE custom_edges (
 tree_id UUID NOT NULL REFERENCES custom_trees(id) ON DELETE CASCADE,
 parent_id UUID NOT NULL REFERENCES custom_entities(id) ON DELETE CASCADE,
 child_id UUID NOT NULL REFERENCES custom_entities(id) ON DELETE CASCADE,
 PRIMARY KEY(parent_id, child_id), UNIQUE(child_id), CHECK(parent_id<>child_id)
);
CREATE INDEX idx_custom_edges_tree ON custom_edges(tree_id);

CREATE TABLE custom_tree_access_emails (
 tree_id UUID NOT NULL REFERENCES custom_trees(id) ON DELETE CASCADE,
 email TEXT NOT NULL, PRIMARY KEY(tree_id,email)
);

CREATE TABLE custom_photos (
 id UUID PRIMARY KEY, entity_id UUID NOT NULL REFERENCES custom_entities(id) ON DELETE CASCADE,
 file_name TEXT NOT NULL, mime_type TEXT NOT NULL, size_bytes BIGINT NOT NULL,
 object_key TEXT NOT NULL UNIQUE, is_avatar BOOLEAN NOT NULL DEFAULT FALSE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_custom_photos_entity ON custom_photos(entity_id);
CREATE UNIQUE INDEX uq_custom_entity_avatar ON custom_photos(entity_id) WHERE is_avatar;
