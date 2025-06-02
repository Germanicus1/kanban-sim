-- +goose Up
-- +goose StatementBegin
ALTER TABLE cards
ADD COLUMN order_index INT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cards
DROP COLUMN order_index;
-- +goose StatementEnd
