-- +goose Up
-- +goose StatementBegin
-- 1) Drop the old all-rows constraint
ALTER TABLE columns
  DROP CONSTRAINT IF EXISTS columns_game_id_order_index_key;

-- 2a) Enforce unique order_index *among top-level* columns
CREATE UNIQUE INDEX columns_game_order_idx
  ON columns(game_id, order_index)
  WHERE parent_id IS NULL;

-- 2b) Enforce unique order_index *per parent* for sub-columns
CREATE UNIQUE INDEX columns_sub_order_idx
  ON columns(game_id, parent_id, order_index)
  WHERE parent_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Roll back to the original single constraint
DROP INDEX IF EXISTS columns_sub_order_idx;
DROP INDEX IF EXISTS columns_game_order_idx;

ALTER TABLE columns
  ADD CONSTRAINT columns_game_id_order_index_key
    UNIQUE (game_id, order_index);
-- +goose StatementEnd
