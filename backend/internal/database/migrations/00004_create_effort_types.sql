-- +goose Up
-- +goose StatementBegin
CREATE TABLE effort_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    order_index INT NOT NULL,
    UNIQUE(game_id, order_index)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE effort_types;
-- +goose StatementEnd
