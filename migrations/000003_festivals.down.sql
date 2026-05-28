ALTER TABLE artists DROP CONSTRAINT IF EXISTS artists_festival_name_starts_at;
ALTER TABLE artists DROP COLUMN IF EXISTS festival_id;
ALTER TABLE artists ADD CONSTRAINT artists_name_starts_at UNIQUE (name, starts_at);

DROP INDEX IF EXISTS idx_festivals_starts_at;
DROP TABLE IF EXISTS festivals;
