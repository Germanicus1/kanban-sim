-- +goose Up
-- +goose StatementBegin
-- rename the old “type” and timestamp columns
ALTER TABLE game_events RENAME COLUMN type TO event_type;
ALTER TABLE game_events RENAME COLUMN inserted_at TO created_at;

-- add the card_id fk and JSON payload
ALTER TABLE game_events
  ADD COLUMN card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
  ADD COLUMN payload JSONB NOT NULL DEFAULT '{}'::jsonb;

-- index by game_id for fast lookups
CREATE INDEX idx_game_events_game_id ON game_events(game_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_game_events_game_id;

ALTER TABLE game_events
  DROP COLUMN payload,
  DROP COLUMN card_id;

ALTER TABLE game_events
  RENAME COLUMN created_at TO inserted_at,
  RENAME COLUMN event_type TO type;
-- +goose StatementEnd
