# Public people HTTP API

The catalogue is mounted at `/api/public-persons`. Public records, events and
media are snapshots; publishing and importing always create independent records
and physical S3 objects.

| Method | Route | Purpose |
| --- | --- | --- |
| POST | `/api/public-persons/` | Create an empty owned public person |
| POST | `/api/public-persons/export` | Publish one person from an owned tree |
| GET | `/api/public-persons/random` | Random catalogue |
| GET | `/api/public-persons/search?q=...` | Search catalogue |
| GET | `/api/public-persons/users/{user_id}` | User's publications |
| GET/PUT/DELETE | `/api/public-persons/{id}` | Read/edit/delete publication |
| POST | `/api/public-persons/{id}/import` | Copy into an existing tree |
| POST | `/api/public-persons/{id}/import-as-tree` | Create a new tree from snapshot |
| POST/GET | `/api/public-persons/{id}/photos` | Upload/list copied media |
| GET/DELETE | `/api/public-persons/{id}/photos/{photo_id}` | Read/delete media |

Existing-tree import accepts `attachment` values `PARENT`, `CHILD` or
`PARTNER`, together with `tree_id` and `attach_to_person_id`.
