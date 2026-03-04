Телеграм-бот - напоминалка. Создает заметку из присланного ему текста и в выбранную дату-время присылает эту заметку пользователю.

# Telegram Reminder Bot

Telegram-бот для создания и управления напоминаниями с веб-интерфейсом для выбора даты и времени.

## Возможности

- ✅ Создание напоминаний из текстовых сообщений
- ✅ Установка даты и времени через веб-приложение
- ✅ Просмотр списка напоминаний (сегодня, неделя, месяц, все)
- ✅ Автоматическая отправка напоминаний в заданное время
- ✅ Поддержка часовых поясов пользователей

## Требования

- Go 1.25.6 или выше
- PostgreSQL
- Telegram Bot Token (получить у [@BotFather](https://t.me/botfather))

## Установка

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd telegram-bot
```

### 2. Настройка переменных окружения

Скопируйте `.env.example` в `.env` и заполните необходимые значения:

```bash
cp .env.example .env
```

Отредактируйте `.env`:

```env
BOT_TOKEN=your_bot_token_from_botfather
TG_WEB_APP_URL=https://your-domain.com
DSN=host=localhost user=postgres password=your_password dbname=reminder_bot port=5432 sslmode=disable

# Настройки веб-сервера (опционально)
SERVER_PORT=3000
USE_TLS=false
TLS_CERT_PATH=/path/to/fullchain.pem
TLS_KEY_PATH=/path/to/privkey.pem
```

### 3. Установка зависимостей

```bash
go mod download
```

### 4. Миграция базы данных

```bash
go run migrations/auto.go
```

### 5. Запуск бота

```bash
go run main.go
```

## Конфигурация

### Переменные окружения

| Переменная       | Обязательная | По умолчанию | Описание                           |
|------------------|--------------|--------------|------------------------------------|
| `BOT_TOKEN`      | Да           | -            | Токен Telegram бота                |
| `DSN`            | Да           | -            | Строка подключения к PostgreSQL    |
| `TG_WEB_APP_URL` | Да           | -            | URL веб-приложения для выбора даты |
| `SERVER_PORT`    | Нет          | 3000         | Порт веб-сервера                   |
| `USE_TLS`        | Нет          | false        | Использовать HTTPS                 |
| `TLS_CERT_PATH`  | Нет*         | -            | Путь к SSL сертификату             |
| `TLS_KEY_PATH`   | Нет*         | -            | Путь к приватному ключу SSL        |

*Обязательны если `USE_TLS=true`

### Настройка для Production

Для запуска в production с HTTPS:

```env
USE_TLS=true
SERVER_PORT=443
TLS_CERT_PATH=/etc/letsencrypt/live/your-domain.com/fullchain.pem
TLS_KEY_PATH=/etc/letsencrypt/live/your-domain.com/privkey.pem
```

## Использование

### Команды бота

- `/today` - Показать напоминания на сегодня
- `/week` - Показать напоминания на неделю
- `/month` - Показать напоминания на месяц
- `/all` - Показать все напоминания

### Создание напоминания

1. Отправьте боту текстовое сообщение
2. Нажмите кнопку "Создать"
3. Нажмите "Добавить дату и время" для выбора времени напоминания
4. Выберите дату и время в веб-приложении

### Запуск в режиме разработки

```bash
# Терминал 1: Запуск бота
go run main.go

# Терминал 2: Разработка frontend
cd datePicker
npm install
npm run dev
```

### Сборка для production

```bash
# Backend
go build -o reminder-bot main.go

# Frontend
cd datePicker
npm run build
```

## Graceful Shutdown

Бот корректно обрабатывает сигналы завершения (SIGINT, SIGTERM):

```bash
# Остановка с graceful shutdown
kill -SIGTERM <pid>
# или
Ctrl+C
```