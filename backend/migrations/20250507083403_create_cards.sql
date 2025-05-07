-- +goose Up
-- +goose StatementBegin
CREATE TABLE cards (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  game_id UUID REFERENCES games(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  card_column TEXT NOT NULL,
  class_of_service TEXT,
  value_estimate TEXT,
  effort_analysis INTEGER,
  effort_development INTEGER,
  effort_test INTEGER,
  selected_day INTEGER,
  deployed_day INTEGER,
  created_at TIMESTAMP DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cards;
-- +goose StatementEnd
