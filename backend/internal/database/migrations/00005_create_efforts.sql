-- +goose Up
-- +goose StatementBegin
CREATE TABLE efforts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    effort_type_id UUID NOT NULL REFERENCES effort_types(id) ON DELETE CASCADE,
    estimate INT NOT NULL DEFAULT 0,
    remaining INT NOT NULL DEFAULT 0,
    actual INT NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE efforts;
-- +goose StatementEnd
