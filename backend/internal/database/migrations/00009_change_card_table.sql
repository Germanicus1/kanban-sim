-- +goose Up
-- +goose StatementBegin
-- 1) Drop the existing text column
ALTER TABLE cards
  DROP COLUMN value_estimate;

-- 2) Add a new integer column in its place (default 0 if desired)
ALTER TABLE cards
  ADD COLUMN value_estimate INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1) Drop the integer column
ALTER TABLE cards
  DROP COLUMN value_estimate;

-- 2) Recreate the original text column
ALTER TABLE cards
  ADD COLUMN value_estimate TEXT;
-- +goose StatementEnd
