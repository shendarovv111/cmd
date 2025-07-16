package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/tictactoe/internal/domain"
)

type GameRepository struct {
	db *pgx.Conn
}

func NewGameRepository(db *pgx.Conn) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) Create(game *domain.Game) error {
	board, err := json.Marshal(game.Board)
	if err != nil {
		return err
	}

	players, err := json.Marshal(game.Players)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO games (id, board, players, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = r.db.Exec(context.Background(), query,
		game.ID, board, players, game.Status, game.CreatedAt, game.UpdatedAt)
	return err
}

func (r *GameRepository) Update(game *domain.Game) error {
	board, err := json.Marshal(game.Board)
	if err != nil {
		return err
	}

	players, err := json.Marshal(game.Players)
	if err != nil {
		return err
	}

	query := `
		UPDATE games
		SET board = $1, players = $2, status = $3, updated_at = $4
		WHERE id = $5
	`
	_, err = r.db.Exec(context.Background(), query,
		board, players, game.Status, game.UpdatedAt, game.ID)
	return err
}

func (r *GameRepository) GetByID(id string) (*domain.Game, error) {
	query := `
		SELECT id, board, players, status, created_at, updated_at
		FROM games
		WHERE id = $1
	`

	var game domain.Game
	var boardJSON, playersJSON []byte

	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&game.ID,
		&boardJSON,
		&playersJSON,
		&game.Status,
		&game.CreatedAt,
		&game.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(boardJSON, &game.Board)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(playersJSON, &game.Players)
	if err != nil {
		return nil, err
	}

	return &game, nil
}

func (r *GameRepository) GetAvailableGames() ([]*domain.Game, error) {
	query := `
		SELECT id, board, players, status, created_at, updated_at
		FROM games
		WHERE status = 'waiting'
	`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []*domain.Game
	for rows.Next() {
		var game domain.Game
		var boardJSON, playersJSON []byte

		err := rows.Scan(
			&game.ID,
			&boardJSON,
			&playersJSON,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(boardJSON, &game.Board)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(playersJSON, &game.Players)
		if err != nil {
			return nil, err
		}

		games = append(games, &game)
	}

	return games, nil
}

func (r *GameRepository) GetActiveGamesByUser(userID string) ([]*domain.Game, error) {
	query := `
		SELECT id, board, players, status, created_at, updated_at
		FROM games
		WHERE status = 'active' AND (
			(players->0->>'ID' = $1) OR (players->1->>'ID' = $1)
		)
	`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []*domain.Game
	for rows.Next() {
		var game domain.Game
		var boardJSON, playersJSON []byte

		err := rows.Scan(
			&game.ID,
			&boardJSON,
			&playersJSON,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(boardJSON, &game.Board)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(playersJSON, &game.Players)
		if err != nil {
			return nil, err
		}

		games = append(games, &game)
	}

	return games, nil
}
