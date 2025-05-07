-- +goose Up
-- +goose StatementBegin

INSERT INTO games (id, day, columns) VALUES
  (gen_random_uuid(), 1, '[
    {"id": "options", "name": "Options"},
    {"id": "selected", "name": "Selected", "wipLimit": 5},
    {
      "id": "analysis",
      "name": "Analysis",
      "wipLimit": 3,
      "subcolumns": [
        { "id": "analysis_in_progress", "name": "In Progress" },
        { "id": "analysis_done", "name": "Done" }
      ]
    },
    {
      "id": "development",
      "name": "Development",
      "wipLimit": 3,
      "subcolumns": [
        { "id": "development_in_progress", "name": "In Progress" },
        { "id": "development_done", "name": "Done" }
      ]
    },
    {"id": "test", "name": "Test", "wipLimit": 3},
    {"id": "ready_to_deploy", "name": "Ready to Deploy"},
    {"id": "deployed", "name": "Deployed"}
  ]'::jsonb);

-- Insert initial cards
INSERT INTO cards (game_id, title, card_column, class_of_service, value_estimate, effort_analysis, effort_development, effort_test, selected_day)
VALUES
  ((SELECT id FROM games LIMIT 1), 'S1', 'ready_to_deploy', 'S', 'medium', 4, 7, 3, 1),
  ((SELECT id FROM games LIMIT 1), 'S2', 'test', 'S', 'high', 3, 5, 8, 1),
  ((SELECT id FROM games LIMIT 1), 'S3', 'development_done', 'S', 'low', 2, 6, 4, 1),
  ((SELECT id FROM games LIMIT 1), 'S4', 'development_done', 'S', 'medium', 6, 4, 7, 1),
  ((SELECT id FROM games LIMIT 1), 'S5', 'development_done', 'S', 'medium', 5, 5, 5, 1),
  ((SELECT id FROM games LIMIT 1), 'S6', 'development_in_progress', 'S', 'very high', 7, 3, 6, 1),
  ((SELECT id FROM games LIMIT 1), 'S7', 'analysis_done', 'S', 'high', 6, 5, 2, 1),
  ((SELECT id FROM games LIMIT 1), 'S8', 'analysis_in_progress', 'S', 'medium', 4, 4, 4, 1),
  ((SELECT id FROM games LIMIT 1), 'S9', 'selected', 'S', 'low', 3, 6, 3, 1),
  ((SELECT id FROM games LIMIT 1), 'S10', 'selected', 'S', 'medium', 2, 7, 5, 1),
  ((SELECT id FROM games LIMIT 1), 'S11', 'options', 'S', 'low', 1, 2, 3, NULL),
  ((SELECT id FROM games LIMIT 1), 'I1', 'options', 'I', 'low', 2, 2, 2, NULL),
  ((SELECT id FROM games LIMIT 1), 'F1', 'options', 'F', 'very high', 5, 5, 5, NULL);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM cards WHERE game_id = (SELECT id FROM games LIMIT 1);
DELETE FROM games WHERE id = (SELECT id FROM games LIMIT 1);

-- +goose StatementEnd
