# Volunteer Rating Platform

Закрытая платформа для волонтёрского сообщества, переписанная на Go. В приложении есть вход по заранее созданным аккаунтам, роли, мероприятия, заявки на участие, отзывы, уведомления, рейтинг SP и админка.

## Что нужно установить

- **Go 1.22 или новее**: <https://go.dev/dl/>
- Git: <https://git-scm.com/downloads>

Проект сейчас не требует внешних Go-библиотек: всё собирается стандартной библиотекой Go.

## Быстрый запуск на Windows

Инструкция рассчитана на обычный ноутбук с Windows 10/11 и PowerShell.

1. Откройте **PowerShell**.
2. Перейдите в папку, где хотите держать проект, например:

   ```powershell
   cd $HOME\Desktop
   ```

3. Склонируйте репозиторий и перейдите в него:

   ```powershell
   git clone <URL_ВАШЕГО_РЕПОЗИТОРИЯ> zashekastiki
   cd zashekastiki
   ```

   Если проект уже скачан архивом, просто распакуйте его и выполните `cd` в папку проекта.

4. Проверьте Go:

   ```powershell
   go version
   ```

   Если команда не найдена, установите Go с сайта выше, закройте PowerShell и откройте его заново.

5. Создайте `.env` из примера:

   ```powershell
   Copy-Item .env.example .env
   ```

6. Создайте демо-данные:

   ```powershell
   go run .\cmd\seed
   ```

7. Запустите сервер:

   ```powershell
   go run .\cmd\server
   ```

8. Откройте в браузере:

   ```text
   http://127.0.0.1:8000/login
   ```

9. Демо-логины после `seed`:

   | Роль | Логин | Пароль |
   | --- | --- | --- |
   | Лидер | `leader` | `Password123` |
   | Организатор | `organizer` | `Password123` |
   | Волонтёр | `volunteer` | `Password123` |

Чтобы остановить сервер, вернитесь в PowerShell и нажмите `Ctrl+C`.

## Запуск на macOS/Linux

```bash
cp .env.example .env
make seed
make run
```

Если `make` не установлен, используйте команды напрямую:

```bash
go run ./cmd/seed
go run ./cmd/server
```

## Проверка, что сервер работает

Пока сервер запущен, откройте второй терминал и выполните:

```powershell
curl http://127.0.0.1:8000/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

## Полезные команды

### Windows PowerShell

```powershell
go test ./...
go run .\cmd\seed
go run .\cmd\server
go run .\cmd\make-admin -username leader -password Password123 -full-name "Главный"
go run .\cmd\make-leader -username organizer
```

### macOS/Linux или Git Bash

```bash
make test
make seed
make run
make make-admin USERNAME=leader PASSWORD=Password123 FULL_NAME="Главный"
make make-leader USERNAME=organizer
```

## Где лежат данные

По умолчанию данные сохраняются в файл:

```text
volunteer_platform.json
```

Файл создаётся автоматически после `go run ./cmd/seed` или после первых изменений в приложении. Для чистого запуска можно остановить сервер, удалить `volunteer_platform.json` и снова выполнить seed.

## Переменные окружения

Файл `.env.example` содержит рабочие значения по умолчанию:

```env
APP_NAME=Volunteer Rating Platform
DATABASE_URL=./volunteer_platform.json
SECRET_KEY=change-me-in-production
SESSION_COOKIE_NAME=volunteer_session
ADDR=:8000
```

- `APP_NAME` — название приложения.
- `DATABASE_URL` — путь к JSON-файлу хранилища. Старый вид `sqlite:///./volunteer_platform.db` тоже принимается и автоматически преобразуется в JSON-файл рядом с проектом.
- `SECRET_KEY` — ключ для cookie-сессий; для реального деплоя обязательно поменяйте.
- `SESSION_COOKIE_NAME` — имя cookie сессии.
- `ADDR` — адрес HTTP-сервера. Для локального запуска обычно достаточно `:8000`.

## Частые проблемы на Windows

### `go` не является внутренней или внешней командой

Go не установлен или PowerShell открыт до установки Go. Установите Go, закройте PowerShell и откройте новый.

### Порт 8000 уже занят

Измените порт в `.env`, например:

```env
ADDR=:8080
```

После этого запускайте сервер и открывайте `http://127.0.0.1:8080/login`.

### Браузер не открывает страницу

Проверьте, что в окне PowerShell есть строка вроде:

```text
listening :8000
```

Если её нет — сервер не запущен или завершился с ошибкой. Скопируйте текст ошибки из PowerShell.

## Структура проекта

Структура приведена к Gopherledger-style: `cmd` содержит только точки входа, а код приложения разделён на конфигурацию, домен, репозиторий, сервисы, HTTP handlers, router, middleware, templates и upload.

- `cmd/server` — минимальная точка входа веб-сервера: загружает config, открывает JSON repository, создаёт services/handlers/router, запускает HTTP server и корректно завершает его по `Ctrl+C`/SIGTERM.
- `cmd/seed` — CLI для демо-данных; использует `internal/config`, `internal/repository`, `internal/service`, `internal/domain`.
- `cmd/make-admin` — CLI для создания/обновления leader-пользователя.
- `cmd/make-leader` — CLI для назначения существующего пользователя лидером.
- `internal/config` — загрузка `.env` и переменных окружения (`APP_NAME`, `DATABASE_URL`, `SECRET_KEY`, `SESSION_COOKIE_NAME`, `ADDR`, пути templates/static/upload).
- `internal/domain` — чистые структуры предметной области, разнесённые по файлам: `user.go`, `event.go`, `review.go`, `rule.go`, `participant.go`, `notification.go`.
- `internal/repository` — JSON-backed CRUD/storage: `repository.go` с интерфейсами, `json_store.go` с `JSONStore`, отдельные файлы для пользователей, мероприятий, отзывов, правил и уведомлений.
- `internal/service` — бизнес-логика без `net/http` и шаблонов: пользователи, пароли, обложки мероприятий, экспорт и дальнейшие бизнес-операции.
- `internal/auth` — password/session/CSRF primitives.
- `internal/handler` — HTTP parsing, redirects, flashes, template rendering и handlers, разнесённые по файлам (`handler.go`, `events.go`, `profile.go`, `admin.go`).
- `internal/router` — сборка HTTP router из handlers.
- `internal/middleware` — security headers, logging, recover и место для reusable auth middleware.
- `internal/templates` — загрузка Go `html/template`; сами шаблоны остаются в `app/templates`.
- `internal/upload` — сохранение multipart/data URL файлов в `app/static/uploads`.
- `internal/platform` — маленький compatibility layer для старых тестов/CLI API; новая разработка должна использовать пакеты выше напрямую.
- `app/templates` — HTML-шаблоны Go.
- `app/static` — CSS, JS, изображения и загруженные файлы.

Правило для дальнейшей разработки: новые сущности сначала описываются в `internal/domain`, CRUD добавляется в `internal/repository`, бизнес-операции — в `internal/service`, затем HTTP-форма/страница подключается в `internal/handler` и маршрут — в `internal/router`.
