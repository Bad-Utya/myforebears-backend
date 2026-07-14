# Custom tree HTTP API

All mutation endpoints require a bearer token. Routes containing `tree_id` use
the same owner/read-access rules as genealogical trees.

| Method | Route | Purpose |
| --- | --- | --- |
| POST | `/api/custom-trees/` | Create a tree and its root entity |
| GET | `/api/custom-trees/` | List the current user's trees |
| GET | `/api/custom-trees/public/random?limit=N` | Random public catalogue |
| GET | `/api/custom-trees/public/search?q=...&limit=N` | Search public trees |
| GET | `/api/custom-trees/public/users/{user_id}` | User's public trees |
| GET/PUT/DELETE | `/api/custom-trees/{tree_id}` | Read, edit or delete a tree |
| GET | `/api/custom-trees/{tree_id}/content` | Tree, entities and edges |
| POST/GET | `/api/custom-trees/{tree_id}/entities` | Create/list entities |
| GET/PUT/DELETE | `/api/custom-trees/{tree_id}/entities/{entity_id}` | Entity CRUD |
| POST/DELETE | `/api/custom-trees/{tree_id}/edges` | Add/remove a directed edge |
| POST/GET/DELETE | `/api/custom-trees/{tree_id}/access-emails` | Shared access |
| POST/GET | `/api/custom-trees/{tree_id}/entities/{entity_id}/photos` | Upload/list media |
| GET/DELETE | `/api/custom-trees/{tree_id}/entities/{entity_id}/photos/{photo_id}` | Read/delete media |
| GET | `/api/custom-trees/{tree_id}/coordinates?root_entity_id=...` | BFS/annealed coordinates |
| GET | `/api/custom-trees/{tree_id}/svg?root_entity_id=...` | Render SVG |

Creating an entity requires `parent_id`. An entity may have only one parent.
Cycles are rejected. The root and any entity with children cannot be deleted.

Tree creation body:

```json
{
  "name": "University",
  "description": "Academic supervision",
  "relation_down": "student",
  "relation_up": "scientific supervisor",
  "root_entity_name": "Professor"
}
```

Photo upload uses multipart form fields `file` and optional `is_avatar=true`.
