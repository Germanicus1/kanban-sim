-- +goose Up
-- +goose StatementBegin
CREATE TYPE column_type AS ENUM ('queue', 'active', 'done');

CREATE TABLE columns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    wip_limit INT DEFAULT 0,
    col_type column_type NOT NULL DEFAULT 'queue',
    parent_id UUID REFERENCES columns(id) ON DELETE CASCADE,
    order_index INT NOT NULL
);

-- Unique index for top‐level columns (parent_id IS NULL)
CREATE UNIQUE INDEX columns_game_order_idx
  ON columns (game_id, order_index)
  WHERE parent_id IS NULL;

-- Unique index for sub‐columns (parent_id IS NOT NULL)
CREATE UNIQUE INDEX columns_sub_order_idx
  ON columns (game_id, parent_id, order_index)
  WHERE parent_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS columns_sub_order_idx;
DROP INDEX IF EXISTS columns_game_order_idx;
DROP TABLE IF EXISTS columns;
DROP TYPE IF EXISTS column_type;
-- +goose StatementEnd
