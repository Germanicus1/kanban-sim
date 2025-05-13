-- +goose Up
-- +goose StatementBegin
CREATE TABLE columns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    parent_id UUID REFERENCES columns(id) ON DELETE CASCADE,
    order_index INT NOT NULL,
    UNIQUE(game_id, order_index)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE columns;
-- +goose StatementEnd
