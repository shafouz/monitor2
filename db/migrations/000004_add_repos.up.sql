CREATE TABLE IF NOT EXISTS Repository (
  id SERIAL PRIMARY KEY,
  url TEXT NOT NULL UNIQUE,
  directory TEXT NOT NULL UNIQUE,
  watched_files TEXT NOT NULL,
  remote TEXT NOT NULL,
  schedule_hours INTEGER NOT NULL DEFAULT 24,
  deleted BOOLEAN NOT NULL DEFAULT false,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_repository_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = current_timestamp;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER repository_updated_at
BEFORE UPDATE
ON repository
FOR EACH ROW
EXECUTE FUNCTION update_repository_updated_at();
