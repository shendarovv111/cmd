package app

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tictactoe/internal/domain"
	"github.com/tictactoe/internal/dto"
)

type GameService struct {
	repo domain.GameRepository
}

func NewGameService(repo domain.GameRepository) *GameService {
	return &GameService{repo: repo}
}

func (s *GameService) CreateGame(req dto.CreateGameRequest) (*dto.OutgoingMessage, error) {
	game := domain.NewGame(req.UserID)
	game.ID = uuid.New().String()

	if err := s.repo.Create(game); err != nil {
		return nil, fmt.Errorf("ошибка создания игры: %w", err)
	}

	return dto.NewOutgoingMessage(
		req.UserID,
		"Игра создана! Ожидаем второго игрока...",
		[]dto.Button{{Text: "Список игр", Action: "/list"}},
	), nil
}

func (s *GameService) ListGames(userID string) (*dto.OutgoingMessage, error) {
	games, err := s.repo.GetAvailableGames()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка игр: %w", err)
	}

	if len(games) == 0 {
		return dto.NewOutgoingMessage(
			userID,
			"Нет доступных игр",
			[]dto.Button{
				{Text: "Создать игру", Action: "/new"},
				{Text: "Моя игра", Action: "/mygame"},
			},
		), nil
	}

	var buttons []dto.Button
	for _, game := range games {
		buttons = append(buttons, dto.Button{
			Text:   fmt.Sprintf("Игра %s (создатель: %s)", game.ID[:8], game.Players[0].ID),
			Action: fmt.Sprintf("/join %s", game.ID),
		})
	}

	return dto.NewOutgoingMessage(userID, "Доступные игры:", buttons), nil
}

func (s *GameService) JoinGame(req dto.JoinGameRequest) (*dto.OutgoingMessages, error) {
	game, err := s.repo.GetByID(req.GameID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения игры: %w", err)
	}

	if err := game.JoinGame(req.UserID); err != nil {
		return nil, err
	}

	if err := s.repo.Update(game); err != nil {
		return nil, fmt.Errorf("ошибка обновления игры: %w", err)
	}

	var messages []dto.OutgoingMessage

	playerMessage := s.getGameMessage(game, req.UserID)
	messages = append(messages, *playerMessage)

	creatorMessage := s.getGameMessage(game, game.Players[0].ID)
	messages = append(messages, *creatorMessage)

	return dto.NewOutgoingMessages(messages...), nil
}

func (s *GameService) ShowHelp(userID string) *dto.OutgoingMessage {
	helpText := `🎮 Добро пожаловать в игру Крестики-нолики!

Доступные команды:
• /new - создать новую игру
• /list - список доступных игр

Как играть:
1. Создайте игру командой /new
2. Поделитесь ссылкой на бота с другом
3. Друг присоединяется к игре через /list
4. Игроки делают ходы по очереди

Координаты ходов:
  1 2 3
A . . .
B . . .
C . . .

Нажимайте на кнопки с координатами (A1, B2, C3 и т.д.) для совершения хода.

🎉 Удачи в игре!`

	return dto.NewOutgoingMessage(
		userID,
		helpText,
		[]dto.Button{
			{Text: "Создать игру", Action: "/new"},
			{Text: "Список игр", Action: "/list"},
			{Text: "Моя игра", Action: "/mygame"},
		},
	)
}

func (s *GameService) getGameMessage(game *domain.Game, userID string) *dto.OutgoingMessage {
	boardText := renderBoard(game.Board)

	var isYourTurn bool
	var yourSymbol string
	var found bool

	for _, p := range game.Players {
		if p.ID == userID {
			isYourTurn = p.IsActive
			yourSymbol = p.Symbol
			found = true
			break
		}
	}

	if !found {
		return dto.NewOutgoingMessage(
			userID,
			"Вы не являетесь участником этой игры",
			nil,
		)
	}

	if game.Status == domain.GameStatusFinished {
		var text string

		if game.CheckWin(yourSymbol) {
			text = "Поздравляем! Вы победили!"
		} else if game.CheckWin(getOpponentSymbol(yourSymbol)) {
			text = "Игра окончена. Победил противник."
		} else {
			text = "Игра окончена. Ничья!"
		}

		return dto.NewOutgoingMessage(
			userID,
			fmt.Sprintf("%s\n\n%s", boardText, text),
			[]dto.Button{
				{Text: "Новая игра", Action: "/new"},
				{Text: "Список игр", Action: "/list"},
			},
		)
	} else if game.Status == domain.GameStatusWaiting {
		return dto.NewOutgoingMessage(
			userID,
			fmt.Sprintf("%s\n\nОжидаем второго игрока...", boardText),
			[]dto.Button{
				{Text: "Список игр", Action: "/list"},
				{Text: "Моя игра", Action: "/mygame"},
			},
		)
	} else if isYourTurn {
		return dto.NewOutgoingMessage(
			userID,
			fmt.Sprintf("%s\n\nВаш ход! Вы играете за %s", boardText, yourSymbol),
			generateMoveButtons(game.ID, game.Board),
		)
	} else {
		return dto.NewOutgoingMessage(
			userID,
			fmt.Sprintf("%s\n\nОжидаем ход противника... Вы играете за %s", boardText, yourSymbol),
			[]dto.Button{
				{Text: "Моя игра", Action: "/mygame"},
			},
		)
	}
}

func (s *GameService) MakeMove(req dto.MakeMoveRequest) (*dto.OutgoingMessages, error) {
	game, err := s.repo.GetByID(req.GameID)
	if err != nil {
		return nil, fmt.Errorf("игра не найдена: %w", err)
	}

	coord, err := parseCoordinate(req.Position)
	if err != nil {
		return nil, fmt.Errorf("неверные координаты: %w", err)
	}

	if err := game.MakeMove(req.UserID, coord); err != nil {
		return nil, err
	}

	if err := s.repo.Update(game); err != nil {
		return nil, fmt.Errorf("ошибка сохранения хода: %w", err)
	}

	var messages []dto.OutgoingMessage

	for _, player := range game.Players {
		if player.ID != "" {
			playerMessage := s.getGameMessage(game, player.ID)
			messages = append(messages, *playerMessage)
		}
	}

	return dto.NewOutgoingMessages(messages...), nil
}

func (s *GameService) ShowGame(req dto.ShowGameRequest) (*dto.OutgoingMessage, error) {
	game, err := s.repo.GetByID(req.GameID)
	if err != nil {
		return nil, fmt.Errorf("игра не найдена: %w", err)
	}

	isPlayer := false
	for _, player := range game.Players {
		if player.ID == req.UserID {
			isPlayer = true
			break
		}
	}

	if !isPlayer {
		return nil, fmt.Errorf("вы не являетесь участником этой игры")
	}

	return s.getGameMessage(game, req.UserID), nil
}

func (s *GameService) GetActiveGame(userID string) (*dto.OutgoingMessage, error) {
	games, err := s.repo.GetActiveGamesByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения игр пользователя: %w", err)
	}

	for _, game := range games {
		if game.Status == domain.GameStatusActive || game.Status == domain.GameStatusWaiting {
			return s.getGameMessage(game, userID), nil
		}
	}

	return dto.NewOutgoingMessage(
		userID,
		"У вас нет активной игры. Создайте новую или присоединитесь к существующей.",
		[]dto.Button{
			{Text: "Создать игру", Action: "/new"},
			{Text: "Список игр", Action: "/list"},
		},
	), nil
}

func (s *GameService) GetGameByID(gameID string) (*domain.Game, error) {
	return s.repo.GetByID(gameID)
}

func (s *GameService) GetGameNotifications(game *domain.Game) *dto.OutgoingMessages {
	var messages []dto.OutgoingMessage

	for _, player := range game.Players {
		if player.ID != "" {
			playerMessage := s.getGameMessage(game, player.ID)
			messages = append(messages, *playerMessage)
		}
	}

	return dto.NewOutgoingMessages(messages...)
}

func renderBoard(board [3][3]string) string {
	var result string
	result += "   1 2 3\n"
	result += "  +-+-+-+\n"
	rows := []string{"A", "B", "C"}

	for i := 0; i < 3; i++ {
		result += rows[i] + " |"
		for j := 0; j < 3; j++ {
			cell := board[i][j]
			if cell == "" {
				cell = "_"
			}
			result += cell + "|"
		}
		result += "\n"
	}

	return result
}

func generateMoveButtons(gameID string, board [3][3]string) []dto.Button {
	var buttons []dto.Button
	rows := []string{"A", "B", "C"}
	cols := []string{"1", "2", "3"}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			coord := rows[i] + cols[j]
			buttons = append(buttons, dto.Button{
				Text:   coord,
				Action: "/move " + gameID + " " + coord,
			})
		}
	}

	return buttons
}

func parseCoordinate(text string) (domain.Coordinate, error) {
	if len(text) != 2 {
		return domain.Coordinate{}, domain.ErrInvalidCoordinate
	}

	rowMap := map[byte]int{'A': 0, 'B': 1, 'C': 2}
	colMap := map[byte]int{'1': 0, '2': 1, '3': 2}

	row, ok := rowMap[text[0]]
	if !ok {
		return domain.Coordinate{}, domain.ErrInvalidCoordinate
	}

	col, ok := colMap[text[1]]
	if !ok {
		return domain.Coordinate{}, domain.ErrInvalidCoordinate
	}

	return domain.Coordinate{Row: row, Column: col}, nil
}

func getOpponentSymbol(symbol string) string {
	if symbol == "X" {
		return "O"
	}
	return "X"
}
