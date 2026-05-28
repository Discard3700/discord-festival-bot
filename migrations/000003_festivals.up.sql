CREATE TABLE festivals (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE,
    starts_at  TIMESTAMPTZ NOT NULL,
    ends_at    TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE artists ADD COLUMN festival_id INT NOT NULL REFERENCES festivals(id) ON DELETE CASCADE;

ALTER TABLE artists DROP CONSTRAINT artists_name_starts_at;
ALTER TABLE artists ADD CONSTRAINT artists_festival_name_starts_at UNIQUE (festival_id, name, starts_at);

CREATE INDEX idx_festivals_starts_at ON festivals (starts_at);
