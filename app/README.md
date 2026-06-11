# Frontend launch

Этот frontend работает с готовым backend из папки `backend`. Один `go run ./cmd/server` запускает только JSON API backend, поэтому страницу нужно запускать отдельной командой из папки `app`.

## Быстрый запуск без npm

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

## Настройки

По умолчанию frontend проксирует API в `http://127.0.0.1:8080` — это порт backend из `backend/config.yaml`.

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

## Альтернативный запуск через Node.js

```bash
cd app
npm install
npm run dev
```

Для production-сборки:

```bash
cd app
npm run build
npm start
```
