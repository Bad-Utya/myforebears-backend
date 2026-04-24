CREATE TABLE IF NOT EXISTS tree_access_emails
(
    tree_id UUID NOT NULL REFERENCES trees (id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tree_id, email)
);

CREATE INDEX IF NOT EXISTS idx_tree_access_emails_tree_id ON tree_access_emails (tree_id);
