-- +goose Up
-- Single-row bookmark of the last Plaid transfer event processed, so each
-- webhook syncs only new events instead of replaying history.
CREATE TABLE IF NOT EXISTS transfer_event_cursor (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    last_event_id INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO transfer_event_cursor (id, last_event_id) VALUES (1, 0)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS transfer_event_cursor;
