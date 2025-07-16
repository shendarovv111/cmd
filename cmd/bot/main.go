package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type IncomingMessage struct {
	UserID string  `json:"userId"`
	Text   *string `json:"text,omitempty"`
	Action *string `json:"action,omitempty"`
}

type OutgoingMessage struct {
	UserID  string   `json:"userId"`
	Text    string   `json:"text"`
	Buttons []Button `json:"buttons"`
}

type OutgoingMessages struct {
	Messages []OutgoingMessage `json:"messages"`
}

type Button struct {
	Text   string `json:"text"`
	Action string `json:"action"`
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("файл .env не найден: %v", err)
	}
}

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Не указан токен бота в BOT_TOKEN")
	}

	serviceURL := os.Getenv("SERVICE_URL")
	if serviceURL == "" {
		serviceURL = "http://localhost:8080/command"
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}

	log.Printf("Бот запущен: @%s", bot.Self.UserName)
	log.Printf("Backend URL: %s", serviceURL)

	u := tgbotapi.NewUpdate(-1)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message, serviceURL)
		} else if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery, serviceURL)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, serviceURL string) {
	userID := fmt.Sprintf("chat_%d", message.Chat.ID)
	text := message.Text

	response, status := sendToBackend(serviceURL, userID, text, nil)
	sendResponse(bot, message.Chat.ID, response)

	if strings.HasPrefix(text, "/join ") || strings.HasPrefix(text, "/move ") {
		if successResponse, ok := response.(OutgoingMessage); ok {
			if status == http.StatusOK && !strings.Contains(successResponse.Text, "ошибка") && !strings.Contains(successResponse.Text, "не ваш ход") && !strings.Contains(successResponse.Text, "не активна") {
				log.Printf("Отправляем push-уведомления для успешной команды: %s", text)
				notifyURL := strings.Replace(serviceURL, "/command", "/notify", 1)
				notifyResponse, _ := sendToBackend(notifyURL, userID, text, nil)

				if messages, ok := notifyResponse.(OutgoingMessages); ok {
					log.Printf("Получено %d push-уведомлений", len(messages.Messages))
					for _, msg := range messages.Messages {
						targetChatID := extractChatIDFromUserID(msg.UserID)
						if targetChatID != 0 && targetChatID != message.Chat.ID {
							log.Printf("Отправляем push-уведомление игроку %d", targetChatID)
							sendSingleMessage(bot, targetChatID, msg)
						}
					}
				} else {
					log.Printf("Не удалось получить push-уведомления для команды: %s", text)
				}
			} else {
				log.Printf("Команда завершилась ошибкой, push-уведомления не отправляем: %s", text)
			}
		}
	}
}

func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, serviceURL string) {
	userID := fmt.Sprintf("chat_%d", callback.Message.Chat.ID)
	action := callback.Data

	response, status := sendToBackend(serviceURL, userID, "", &action)
	sendResponse(bot, callback.Message.Chat.ID, response)

	if strings.HasPrefix(action, "/join ") || strings.HasPrefix(action, "/move ") {
		if successResponse, ok := response.(OutgoingMessage); ok {
			if status == http.StatusOK && !strings.Contains(successResponse.Text, "ошибка") && !strings.Contains(successResponse.Text, "не ваш ход") && !strings.Contains(successResponse.Text, "не активна") {
				log.Printf("Отправляем push-уведомления для успешного действия: %s", action)
				notifyURL := strings.Replace(serviceURL, "/command", "/notify", 1)
				notifyResponse, _ := sendToBackend(notifyURL, userID, "", &action)

				if messages, ok := notifyResponse.(OutgoingMessages); ok {
					log.Printf("Получено %d push-уведомлений для действия", len(messages.Messages))
					for _, msg := range messages.Messages {
						targetChatID := extractChatIDFromUserID(msg.UserID)
						if targetChatID != 0 && targetChatID != callback.Message.Chat.ID {
							log.Printf("Отправляем push-уведомление игроку %d для действия", targetChatID)
							sendSingleMessage(bot, targetChatID, msg)
						}
					}
				} else {
					log.Printf("Не удалось получить push-уведомления для действия: %s", action)
				}
			} else {
				log.Printf("Действие завершилось ошибкой, push-уведомления не отправляем: %s", action)
			}
		}
	}

	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	bot.Send(callbackConfig)
}

func sendToBackend(serviceURL, userID, text string, action *string) (interface{}, int) {
	msg := IncomingMessage{UserID: userID}
	if text != "" {
		msg.Text = &text
	}
	if action != nil {
		msg.Action = action
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Ошибка при сериализации JSON: %v", err)
		return OutgoingMessage{
			UserID: userID,
			Text:   "Произошла ошибка при обработке запроса",
		}, http.StatusInternalServerError
	}

	resp, err := http.Post(serviceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Ошибка при отправке запроса: %v", err)
		return OutgoingMessage{
			UserID: userID,
			Text:   "Не удалось подключиться к сервису",
		}, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка при чтении ответа: %v", err)
		return OutgoingMessage{
			UserID: userID,
			Text:   "Ошибка при чтении ответа от сервиса",
		}, http.StatusInternalServerError
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка от сервера: %s", body)
		return OutgoingMessage{
			UserID: userID,
			Text:   "Сервер вернул ошибку: " + string(body),
		}, resp.StatusCode
	}

	var outMessages OutgoingMessages
	if err := json.Unmarshal(body, &outMessages); err == nil && len(outMessages.Messages) > 0 {
		return outMessages, resp.StatusCode
	}

	var outMsg OutgoingMessage
	if err := json.Unmarshal(body, &outMsg); err != nil {
		log.Printf("Ошибка при десериализации ответа: %v", err)
		return OutgoingMessage{
			UserID: userID,
			Text:   "Ошибка при обработке ответа от сервиса",
		}, http.StatusInternalServerError
	}

	return outMsg, resp.StatusCode
}

func sendResponse(bot *tgbotapi.BotAPI, chatID int64, response interface{}) {
	switch resp := response.(type) {
	case OutgoingMessages:
		for _, msg := range resp.Messages {
			targetChatID := extractChatIDFromUserID(msg.UserID)
			if targetChatID != 0 {
				sendSingleMessage(bot, targetChatID, msg)
			} else {
				log.Printf("Не удалось извлечь chatID из userID: %s", msg.UserID)
			}
		}
	case OutgoingMessage:
		sendSingleMessage(bot, chatID, resp)
	default:
		log.Printf("Неизвестный тип ответа: %T", response)
	}
}

func extractChatIDFromUserID(userID string) int64 {
	if strings.HasPrefix(userID, "chat_") {
		chatIDStr := strings.TrimPrefix(userID, "chat_")
		if chatID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
			return chatID
		}
	}
	return 0
}

func sendSingleMessage(bot *tgbotapi.BotAPI, chatID int64, response OutgoingMessage) {
	msg := tgbotapi.NewMessage(chatID, response.Text)

	if len(response.Buttons) > 0 {
		var rows [][]tgbotapi.InlineKeyboardButton
		var row []tgbotapi.InlineKeyboardButton

		for i, btn := range response.Buttons {
			button := tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.Action)
			row = append(row, button)

			if (i+1)%3 == 0 || i == len(response.Buttons)-1 {
				rows = append(rows, row)
				row = []tgbotapi.InlineKeyboardButton{}
			}
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
		msg.ReplyMarkup = keyboard
	}

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}
