ALTER TABLE artists ADD CONSTRAINT artists_name_starts_at UNIQUE (name, starts_at);
