package domain

import (
	"time"
)

type Player struct {
	ID       string
	Symbol   string
	IsActive bool
}

type Game struct {
	ID        string
	Board     [3][3]string
	Players   [2]Player
	Status    GameStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GameStatus string

const (
	GameStatusWaiting  GameStatus = "waiting"
	GameStatusActive   GameStatus = "active"
	GameStatusFinished GameStatus = "finished"
)

type Coordinate struct {
	Row    int
	Column int
}

func NewGame(creatorID string) *Game {
	return &Game{
		Status:    GameStatusWaiting,
		Players:   [2]Player{{ID: creatorID, IsActive: false}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (g *Game) MakeMove(playerID string, coord Coordinate) error {
	if g.Status != GameStatusActive {
		return ErrGameNotActive
	}

	var player *Player
	for i := range g.Players {
		if g.Players[i].IsActive && g.Players[i].ID == playerID {
			player = &g.Players[i]
			break
		}
	}
	if player == nil {
		return ErrNotPlayerTurn
	}

	if coord.Row < 0 || coord.Row >= 3 || coord.Column < 0 || coord.Column >= 3 ||
		g.Board[coord.Row][coord.Column] != "" {
		return ErrInvalidMove
	}

	g.Board[coord.Row][coord.Column] = player.Symbol
	g.UpdatedAt = time.Now()

	gameWon := false
	for _, p := range g.Players {
		if g.CheckWin(p.Symbol) {
			gameWon = true
			break
		}
	}

	if gameWon || g.isBoardFull() {
		g.Status = GameStatusFinished
	} else {
		g.Players[0].IsActive = !g.Players[0].IsActive
		g.Players[1].IsActive = !g.Players[1].IsActive
	}

	return nil
}

func (g *Game) JoinGame(playerID string) error {
	if g.Status != GameStatusWaiting {
		return ErrCannotJoin
	}

	if g.Players[0].ID == playerID {
		return ErrAlreadyInGame
	}

	g.Players[1] = Player{ID: playerID}
	g.Status = GameStatusActive

	if time.Now().UnixNano()%2 == 0 {
		g.Players[0].Symbol = "X"
		g.Players[1].Symbol = "O"
		g.Players[0].IsActive = true
		g.Players[1].IsActive = false
	} else {
		g.Players[0].Symbol = "O"
		g.Players[1].Symbol = "X"
		g.Players[0].IsActive = false
		g.Players[1].IsActive = true
	}

	return nil
}

func (g *Game) CheckWin(symbol string) bool {
	winPatterns := [][3][2]int{
		{{0, 0}, {0, 1}, {0, 2}},
		{{1, 0}, {1, 1}, {1, 2}},
		{{2, 0}, {2, 1}, {2, 2}},
		{{0, 0}, {1, 0}, {2, 0}},
		{{0, 1}, {1, 1}, {2, 1}},
		{{0, 2}, {1, 2}, {2, 2}},
		{{0, 0}, {1, 1}, {2, 2}},
		{{0, 2}, {1, 1}, {2, 0}},
	}

	for _, pattern := range winPatterns {
		win := true
		for _, pos := range pattern {
			if g.Board[pos[0]][pos[1]] != symbol {
				win = false
				break
			}
		}
		if win {
			return true
		}
	}

	return false
}

func (g *Game) isBoardFull() bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.Board[i][j] == "" {
				return false
			}
		}
	}
	return true
}

func (g *Game) GetActivePlayer() *Player {
	for i := range g.Players {
		if g.Players[i].IsActive {
			return &g.Players[i]
		}
	}
	return nil
}
