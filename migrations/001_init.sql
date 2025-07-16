-- +goose Up
CREATE TABLE games (
    id VARCHAR(36) PRIMARY KEY,
    board JSONB NOT NULL,
    players JSONB NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_games_status ON games(status);

-- +goose Down
DROP TABLE games; 