package domain

import "errors"

var (
	ErrGameNotActive = errors.New("игра не активна")
	ErrGameNotFound  = errors.New("игра не найдена")
	ErrGameFinished  = errors.New("игра завершена")

	ErrNotPlayerTurn = errors.New("сейчас не ваш ход")
	ErrAlreadyInGame = errors.New("вы уже в игре")
	ErrCannotJoin    = errors.New("к этой игре нельзя присоединиться")

	ErrInvalidMove       = errors.New("недопустимый ход")
	ErrInvalidCoordinate = errors.New("неверные координаты")
)
