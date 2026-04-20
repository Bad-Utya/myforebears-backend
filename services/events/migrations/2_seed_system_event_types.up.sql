INSERT INTO event_types (id, owner_user_id, is_system, name, primary_persons_mode, primary_persons_count)
VALUES
    ('4af8f935-180f-4be6-8f7a-f6ecf90af4b2', 0, TRUE, 'Birth', 'FIXED', 1),
    ('7e92347c-b30d-474e-abdd-48f62cb0f6cf', 0, TRUE, 'Death', 'FIXED', 1),
    ('2c2f5f12-5476-4ef4-89df-85d0a6f4a6bc', 0, TRUE, 'Marriage', 'FIXED', 2),
    ('f1df881e-d7ec-4e97-9f47-4fd926e49532', 0, TRUE, 'Divorce', 'FIXED', 2),
    ('0f6538f2-c06f-4fe1-bd70-412f76db4fda', 0, TRUE, 'House Purchase', 'UNLIMITED', 0),
    ('17cbf04f-279b-4c35-985f-49cfdb4fa88f', 0, TRUE, 'Pet Purchase', 'UNLIMITED', 0)
ON CONFLICT (id) DO NOTHING;
