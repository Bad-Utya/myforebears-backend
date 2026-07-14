CREATE TABLE IF NOT EXISTS public_persons
(
    id              UUID PRIMARY KEY,
    owner_user_id   INT NOT NULL,
    first_name      TEXT NOT NULL DEFAULT '',
    last_name       TEXT NOT NULL DEFAULT '',
    patronymic      TEXT NOT NULL DEFAULT '',
    gender          gender_enum,
    biography       TEXT NOT NULL DEFAULT '',
    avatar_photo_id UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_public_persons_owner ON public_persons (owner_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_public_persons_search ON public_persons
    USING GIN (to_tsvector('simple', first_name || ' ' || last_name || ' ' || patronymic || ' ' || biography));

CREATE TABLE IF NOT EXISTS public_person_events
(
    id                 UUID PRIMARY KEY,
    public_person_id   UUID NOT NULL REFERENCES public_persons(id) ON DELETE CASCADE,
    source_event_id    UUID,
    event_type_id      UUID,
    event_type_name    TEXT NOT NULL,
    date_iso           TEXT NOT NULL DEFAULT '',
    date_precision     TEXT NOT NULL DEFAULT 'DAY',
    date_bound         TEXT NOT NULL DEFAULT 'EXACT',
    date_unknown       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_public_person_events_person ON public_person_events (public_person_id);
