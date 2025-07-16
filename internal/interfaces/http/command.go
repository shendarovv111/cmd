package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tictactoe/internal/app"
	"github.com/tictactoe/internal/domain"
	"github.com/tictactoe/internal/dto"
)

type CommandHandler struct {
	gameService *app.GameService
}

func NewCommandHandler(gameService *app.GameService) *CommandHandler {
	return &CommandHandler{gameService: gameService}
}

func (h *CommandHandler) RegisterRoutes(r chi.Router) {
	r.Post("/command", h.HandleCommand)
	r.Post("/notify", h.HandleNotify)
}

func (h *CommandHandler) HandleCommand(w http.ResponseWriter, r *http.Request) {
	var msg dto.IncomingMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "неверный формат сообщения", http.StatusBadRequest)
		return
	}

	command := h.getCommand(msg)
	if command == "" {
		http.Error(w, "отсутствует команда", http.StatusBadRequest)
		return
	}

	response, err := h.executeCommand(command, msg.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch resp := response.(type) {
	case *dto.OutgoingMessages:
		json.NewEncoder(w).Encode(resp)
	case *dto.OutgoingMessage:
		json.NewEncoder(w).Encode(resp)
	default:
		json.NewEncoder(w).Encode(response)
	}
}

func (h *CommandHandler) HandleNotify(w http.ResponseWriter, r *http.Request) {
	var msg dto.IncomingMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "неверный формат сообщения", http.StatusBadRequest)
		return
	}

	command := h.getCommand(msg)
	if command == "" {
		http.Error(w, "отсутствует команда", http.StatusBadRequest)
		return
	}

	var response interface{}
	var err error

	switch {
	case strings.HasPrefix(command, "/join "):
		gameID := strings.TrimPrefix(command, "/join ")
		game, err := h.gameService.GetGameByID(gameID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response = h.gameService.GetGameNotifications(game)

	case strings.HasPrefix(command, "/move "):
		parts := strings.Split(command, " ")
		if len(parts) != 3 {
			http.Error(w, domain.ErrInvalidMove.Error(), http.StatusBadRequest)
			return
		}
		gameID := parts[1]
		game, err := h.gameService.GetGameByID(gameID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response = h.gameService.GetGameNotifications(game)

	default:
		http.Error(w, "команда не поддерживает push-уведомления", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch resp := response.(type) {
	case *dto.OutgoingMessages:
		json.NewEncoder(w).Encode(resp)
	case *dto.OutgoingMessage:
		messages := dto.NewOutgoingMessages(*resp)
		json.NewEncoder(w).Encode(messages)
	default:
		json.NewEncoder(w).Encode(response)
	}
}

func (h *CommandHandler) getCommand(msg dto.IncomingMessage) string {
	if msg.Text != nil {
		return strings.TrimSpace(*msg.Text)
	}
	if msg.Action != nil {
		return strings.TrimSpace(*msg.Action)
	}
	return ""
}

func (h *CommandHandler) executeCommand(command, userID string) (interface{}, error) {
	switch {
	case command == "/new":
		return h.gameService.CreateGame(dto.CreateGameRequest{UserID: userID})

	case command == "/list":
		return h.gameService.ListGames(userID)

	case command == "/start":
		return h.gameService.ShowHelp(userID), nil

	case command == "/help", command == "":
		return h.gameService.ShowHelp(userID), nil

	case strings.HasPrefix(command, "/join "):
		gameID := strings.TrimPrefix(command, "/join ")
		response, err := h.gameService.JoinGame(dto.JoinGameRequest{
			UserID: userID,
			GameID: gameID,
		})
		if err != nil {
			return nil, err
		}
		for _, msg := range response.Messages {
			if msg.UserID == userID {
				return &msg, nil
			}
		}
		return nil, fmt.Errorf("сообщение для пользователя не найдено")

	case strings.HasPrefix(command, "/move "):
		parts := strings.Split(command, " ")
		if len(parts) != 3 {
			return nil, domain.ErrInvalidMove
		}
		response, err := h.gameService.MakeMove(dto.MakeMoveRequest{
			UserID:   userID,
			GameID:   parts[1],
			Position: parts[2],
		})
		if err != nil {
			return nil, err
		}
		for _, msg := range response.Messages {
			if msg.UserID == userID {
				return &msg, nil
			}
		}
		return nil, fmt.Errorf("сообщение для пользователя не найдено")

	case strings.HasPrefix(command, "/game "):
		gameID := strings.TrimPrefix(command, "/game ")
		return h.gameService.ShowGame(dto.ShowGameRequest{
			UserID: userID,
			GameID: gameID,
		})

	case command == "/mygame":
		return h.gameService.GetActiveGame(userID)

	default:
		return h.gameService.ShowHelp(userID), nil
	}
}
