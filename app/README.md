# Frontend на Go

Frontend собран как отдельное Go-приложение в папке `app`. Он похож на старую серверную версию: обычные страницы, формы и редиректы, а не отдельная SPA. Все данные и действия берутся из готового backend API в папке `backend`.

Важно: один `go run ./cmd/server` запускает только backend API. Чтобы открыть сайт в браузере, нужно запустить frontend отдельной командой.

## Быстрый запуск

Терминал 1 — backend:

```bash
cd backend
go run ./cmd/server
```

Терминал 2 — frontend:

```bash
cd app
go run .
```

Откройте сайт:

```text
http://127.0.0.1:5173
```

## Что умеет frontend

- вход и регистрация через `/auth/login` и `/auth/register`;
- хранение backend token в HttpOnly cookie frontend-сервера;
- профиль через `/auth/me`;
- список ивентов и участников через `/events` и `/events/{id}/participants`;
- запись и отмена записи через `/events/{id}/register`;
- создание ивента для `Admin`/`Organizer` через `/events`;
- админские действия, которые реально есть в backend: promote user и approve finished event;
- страницы «Мерч», «Наши возможности», «Поиск» в стиле старого интерфейса.

## Настройки

По умолчанию frontend ходит в backend по адресу `http://127.0.0.1:8080`.

Если backend запущен по другому адресу:

```bash
cd app
BACKEND_URL=http://127.0.0.1:8080 go run .
```

Если нужен другой адрес frontend:

```bash
cd app
FRONTEND_ADDR=127.0.0.1:3000 go run .
```
