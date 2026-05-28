CREATE TABLE artists (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    stage       TEXT NOT NULL,
    starts_at   TIMESTAMPTZ NOT NULL,
    ends_at     TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_artists_starts_at ON artists (starts_at);

CREATE TABLE reminders (
    id          SERIAL PRIMARY KEY,
    artist_id   INT NOT NULL REFERENCES artists (id) ON DELETE CASCADE,
    channel_id  TEXT NOT NULL,
    remind_at   TIMESTAMPTZ NOT NULL,
    sent        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reminders_remind_at ON reminders (remind_at) WHERE NOT sent;

CREATE TABLE checklist_items (
    id          SERIAL PRIMARY KEY,
    label       TEXT NOT NULL,
    checked     BOOLEAN NOT NULL DEFAULT FALSE,
    added_by    TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
