package dto

type CreateGameRequest struct {
	UserID   string
	UserName string
}

type JoinGameRequest struct {
	UserID   string
	UserName string
	GameID   string
}

type MakeMoveRequest struct {
	UserID   string
	UserName string
	GameID   string
	Position string
}

type ShowGameRequest struct {
	UserID   string
	UserName string
	GameID   string
}
