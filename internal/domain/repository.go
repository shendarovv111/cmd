package domain

type GameRepository interface {
	Create(game *Game) error
	Update(game *Game) error
	GetByID(id string) (*Game, error)
	GetAvailableGames() ([]*Game, error)
	GetActiveGamesByUser(userID string) ([]*Game, error)
}
