# Tic-Tac-Toe Telegram Bot

TG бот для игры в крестики-нолики.

## быстрый запуск

### 1. склонить проект
```bash
git clone <repository-url>
cd cmd
```
# добавить .env в корень проекта! по примеру .env.exemple

### 2. Запуск проекта
```bash
#остановить процессы если есть 
docker-compose down -v

# сборка и запуск
docker-compose build --no-cache
docker-compose up -d
```

## Управление

```bash
# просмотр логов
docker-compose logs -f

# остановка
docker-compose down

# перезапуск
docker-compose restart
```
