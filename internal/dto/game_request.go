package dto

type CreateGameRequest struct {
	UserID string
}

type JoinGameRequest struct {
	UserID string
	GameID string
}

type MakeMoveRequest struct {
	UserID   string
	GameID   string
	Position string
}

type ShowGameRequest struct {
	UserID string
	GameID string
}
