CREATE CONSTRAINT person_id_unique IF NOT EXISTS
FOR (p:Person)
REQUIRE p.id IS UNIQUE;

CREATE INDEX person_tree_id_idx IF NOT EXISTS
FOR (p:Person)
ON (p.tree_id);
