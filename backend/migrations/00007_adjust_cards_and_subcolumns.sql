-- +goose Up
-- +goose StatementBegin
ALTER TABLE cards
    ADD COLUMN column_order INT DEFAULT 0;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_card_column ON cards(card_column);
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE cards SET column_order = 0 WHERE card_column = 'options';
UPDATE cards SET column_order = 1 WHERE card_column = 'selected';
UPDATE cards SET column_order = 2 WHERE card_column = 'analysis_in_progress';
UPDATE cards SET column_order = 3 WHERE card_column = 'analysis_done';
UPDATE cards SET column_order = 4 WHERE card_column = 'development_in_progress';
UPDATE cards SET column_order = 5 WHERE card_column = 'development_done';
UPDATE cards SET column_order = 6 WHERE card_column = 'test';
UPDATE cards SET column_order = 7 WHERE card_column = 'ready_to_deploy';
UPDATE cards SET column_order = 8 WHERE card_column = 'deployed';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cards DROP COLUMN column_order;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_card_column;
-- +goose StatementEnd