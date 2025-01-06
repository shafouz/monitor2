CREATE TABLE IF NOT EXISTS Endpoint (
  id SERIAL PRIMARY KEY,
  url TEXT NOT NULL UNIQUE,
  status_code INTEGER NOT NULL DEFAULT 0,
  response_body TEXT,
  previous_response_body TEXT,
  schedule_hours INTEGER NOT NULL DEFAULT 8,
  deleted BOOLEAN NOT NULL DEFAULT false,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_endpoint_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = current_timestamp;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER endpoint_updated_at
BEFORE UPDATE
ON endpoint
FOR EACH ROW
EXECUTE FUNCTION update_endpoint_updated_at();
