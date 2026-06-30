# Volunteer Platform API

Платформа для автоматизации волонтёрской деятельности университета.

## 🏗 Архитектура

```
VolunteerPlatform/                    ← корень монорепо
├── README.md                         ← общая документация
├── .gitignore                        ← общий gitignore
│
├── backend/                          ← Go бэкенд
│   ├── cmd/
│   │   └── server/
│   │       └── main.go               ← точка входа
│   ├── internal/
│   │   ├── auth/                     ← JWT токены
│   │   ├── config/                   ← загрузка YAML конфига
│   │   ├── domain/                   ← бизнес-модели (User, Event, Reward...)
│   │   ├── dto/                      ← DTO для API ответов
│   │   ├── handler/                  ← HTTP хэндлеры
│   │   ├── middleware/               ← Auth, CORS, Logging
│   │   ├── router/                   ← маршрутизация
│   │   ├── service/                  ← бизнес-логика
│   │   ├── store/                    ← хранилище (in-memory CRUD)
│   │   └── worker/                   ← фоновые задачи (смена статусов)
│   ├── app/
│   │   └── static/
│   │       └── images/               ← картинки для наград и ивентов
│   │           ├── rewards/          ← hoodie.png, sticker_pack.png...
│   │           └── events/           ← subbotnik.jpg, conference.jpg...
│   ├── .env                          ← секреты (JWT_SECRET, ADMIN_LOGIN...)
│   ├── config.yaml                   ← настройки сервера
│   ├── go.mod
│   └── go.sum
│
└── frontend/                         ← Angular фронтенд
    ├── src/
    │   ├── app/
    │   │   ├── components/           ← компоненты UI
    │   │   ├── services/             ← API-клиенты
    │   │   ├── guards/               ← auth guards
    │   │   ├── models/               ← TypeScript интерфейсы
    │   │   └── app.component.ts
    │   ├── assets/                   ← статика фронта (лого, иконки)
    │   ├── environments/             ← dev/prod конфиги
    │   ├── index.html
    │   ├── main.ts
    │   └── styles.css
    ├── angular.json
    ├── package.json
    ├── tsconfig.json
    └── README.md                     ← документация для фронта

```

## 🔐 Аутентификация

**JWT токены** в заголовке `Authorization`:
```
Authorization: <token>
```

Токен получается после `POST /auth/login` или `POST /auth/register`.

Срок жизни: 24 часа (настраивается в `.env` → `JWT_EXPIRATION_HOURS`).

## 📦 Формат ответов

### Успех:
```json
{
  "success": true,
  "data": { ... }
}
```

### Ошибка:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Человеческое описание"
  }
}
```

## 👥 Роли

| Роль | Что может |
|------|-----------|
| **Volunteer** | Регистрироваться на ивенты, покупать награды, смотреть свой профиль |
| **Organizer** | + Создавать ивенты, подтверждать посещаемость СВОИХ ивентов, банить юзеров |
| **Admin** | + Всё: управление ролями, начисление SP, подтверждение ивентов, создание наград/шаблонов |

## 🌐 API Endpoints

### Публичные (без авторизации)

#### Регистрация
```http
POST /auth/register
Content-Type: application/json

{
  "login": "ivanov",
  "password": "secure123",
  "first_name": "Иван",
  "last_name": "Иванов",
  "middle_name": "Иванович",
  "telegram": "@ivanov"
}

→ 201 Created
{
  "success": true,
  "data": { "token": "eyJhbGc..." }
}
```

#### Логин
```http
POST /auth/login
Content-Type: application/json

{
  "login": "ivanov",
  "password": "secure123"
}

→ 200 OK
{
  "success": true,
  "data": { "token": "eyJhbGc..." }
}
```

### Защищённые (нужен Auth)

#### Мой профиль
```http
GET /auth/me
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": {
    "id": 5,
    "login": "ivanov",
    "first_name": "Иван",
    "last_name": "Иванов",
    "middle_name": "Иванович",
    "telegram": "@ivanov",
    "role": "Volunteer",
    "skill_points": 320
  }
}
```

#### Список ивентов
```http
GET /events
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": {
    "events": [
      {
        "id": 1,
        "title": "Субботник",
        "description": "...",
        "location": "Парк",
        "cover_image_url": "/static/images/events/subbotnik.jpg",
        "status": "EVENT-RECRUITING",
        "start_date": "2026-07-01T10:00:00Z",
        "end_date": "2026-07-01T14:00:00Z",
        "registration_deadline": "2026-06-30T23:59:00Z",
        "max_participants": 50,
        "reserve_participants": 10,
        "skill_points": 50,
        "created_by_id": 2,
        "participants_count": 30,
        "reserve_count": 5
      }
    ],
    "count": 1
  }
}
```

**Статусы ивентов:**
- `EVENT-RECRUITING` — идёт набор
- `EVENT-ACTIVE` — ивент идёт
- `EVENT-FINISHED` — ивент закончился, ждёт подтверждения админом
- `EVENT-CLOSED` — баллы начислены
- `EVENT-CANCELLED` — отменён

#### Регистрация на ивент
```http
POST /events/{id}/register
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": { "message": "Вы успешно зарегистрированы на мероприятие" }
}
```

#### Отмена регистрации
```http
DELETE /events/{id}/register
Authorization: <token>

→ 200 OK
```

#### Участники ивента
```http
GET /events/{id}/participants
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": {
    "event_id": 1,
    "participants": [
      {
        "user_id": 5,
        "first_name": "Иван",
        "last_name": "Иванов",
        "middle_name": "Иванович",
        "telegram": "@ivanov",
        "attendance_confirmed": true
      }
    ],
    "count": 1
  }
}
```

#### Моя история ивентов
```http
GET /users/me/events
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": {
    "user_id": 5,
    "events": [
      {
        "event_id": 1,
        "title": "Субботник",
        "status": "EVENT-CLOSED",
        "joined_at": "2026-06-25T10:00:00Z",
        "attendance_confirmed": true,
        "skill_points": 50
      }
    ],
    "count": 1
  }
}
```

#### История транзакций SP
```http
GET /users/{id}/skill-points/history
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": {
    "user_id": 5,
    "transactions": [
      {
        "id": 1,
        "user_id": 5,
        "points": 50,
        "type": "event",
        "reason": "Участие в мероприятии: Субботник",
        "event_id": 1,
        "created_at": "2026-07-01T15:00:00Z"
      },
      {
        "id": 2,
        "user_id": 5,
        "points": -500,
        "type": "reward",
        "reason": "Получение награды: Худи",
        "event_id": null,
        "created_at": "2026-07-02T10:00:00Z"
      }
    ],
    "count": 2
  }
}
```

**Типы транзакций:**
- `manual` — ручное начисление админом
- `event` — автоматическое за ивент
- `reward` — списание за награду

### Каталог наград (Battle Pass)

#### Все награды (каталог)
```http
GET /admin/rewards
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Худи",
      "description": "Лимитированная коллекция",
      "cost": 500,
      "image_url": "/static/images/rewards/hoodie.png"
    }
  ]
}
```

#### Мои награды (с флагами доступности)
```http
GET /users/me/rewards
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": {
    "user_id": 5,
    "rewards": [
      {
        "reward_id": 1,
        "name": "Худи",
        "description": "Лимитированная коллекция",
        "cost": 500,
        "image_url": "/static/images/rewards/hoodie.png",
        "available": true,
        "claimed": false,
        "picked_up": false
      }
    ]
  }
}
```

**Флаги:**
- `available` — хватает SP для покупки (`skill_points >= cost`)
- `claimed` — уже куплено
- `picked_up` — уже получено физически

#### Купить награду
```http
POST /users/me/rewards/{id}/claim
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": { "message": "Награда забронирована" }
}
```

⚠️ При покупке SP **списываются**!

#### Доступные картинки
```http
GET /assets/images
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": {
    "images": [
      "/static/images/rewards/hoodie.png",
      "/static/images/events/subbotnik.jpg"
    ],
    "count": 2
  }
}
```

### Оргские роуты (Organizer / Admin)

#### Подтвердить посещаемость участника
```http
POST /organizer/events/{event_id}/users/{user_id}/confirm
Authorization: <token>

→ 200 OK
```

⚠️ Только создатель ивента может подтверждать!

#### Забанить юзера
```http
POST /organizer/blacklist
Authorization: <token>
Content-Type: application/json

{ "user_id": 5 }

→ 200 OK
```

Забаненный юзер **не видит** ивенты этого орга в списке.

#### Разбанить
```http
DELETE /organizer/blacklist/{user_id}
Authorization: <token>

→ 200 OK
```

### Админские роуты (только Admin)

#### Поиск юзеров
```http
GET /admin/users?first_name=Иван&last_name=Иванов
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": {
    "users": [
      {
        "id": 5,
        "login": "ivanov",
        "first_name": "Иван",
        "last_name": "Иванов",
        "middle_name": "Иванович",
        "telegram": "@ivanov",
        "role": "Volunteer",
        "skill_points": 320
      }
    ],
    "count": 1
  }
}
```

#### 🎁 Купленные награды юзера (для выдачи мерча)
```http
GET /admin/users/{id}/rewards
Authorization: <token>

→ 200 OK
{
  "success": true,
  "data": {
    "user_id": 5,
    "rewards": [
      {
        "reward_id": 1,
        "name": "Худи",
        "cost": 500,
        "image_url": "/static/images/rewards/hoodie.png",
        "claimed": true,
        "picked_up": false
      }
    ]
  }
}
```

💡 Возвращает **только купленные** награды. Фронт фильтрует `picked_up: false` для списка "к выдаче".

#### Отметить награду как выданную
```http
POST /admin/users/{user_id}/rewards/{reward_id}/pickup
Authorization: <token>

→ 200 OK
```

#### Повысить роль
```http
POST /admin/users/{id}/promote
Authorization: <token>
Content-Type: application/json

{ "role": "Organizer" }

→ 200 OK
```

Возможные роли: `Organizer`, `Admin`.

#### Начислить SP вручную
```http
POST /admin/users/{id}/award
Authorization: <token>
Content-Type: application/json

{
  "points": 100,
  "reason": "Организация конференции"
}

→ 200 OK
```

#### Подтвердить ивент и начислить SP участникам
```http
POST /admin/events/{id}/approve
Authorization: <token>

→ 200 OK
```

⚠️ Работает только для ивентов в статусе `EVENT-FINISHED`. SP начисляются только тем, у кого `attendance_confirmed: true`.

#### Создать награду
```http
POST /admin/rewards
Authorization: <token>
Content-Type: application/json

{
  "name": "Худи",
  "description": "Лимитированная коллекция",
  "cost": 500,
  "image_url": "/static/images/rewards/hoodie.png"
}

→ 201 Created
```

⚠️ `image_url` должен быть из списка `GET /assets/images`.

#### Создать шаблон ивента
```http
POST /admin/event-templates
Authorization: <token>
Content-Type: application/json

{
  "title": "Субботник",
  "description": "Уборка парка",
  "location": "Парк",
  "image_url": "/static/images/events/subbotnik.jpg",
  "duration_minutes": 240,
  "max_participants": 50,
  "reserve_participants": 10,
  "skill_points": 50
}

→ 201 Created
```

#### Список шаблонов
```http
GET /admin/event-templates
Authorization: <token>

→ 200 OK
```

#### Создать ивент (Org/Admin)
```http
POST /events
Authorization: <token>
Content-Type: application/json

{
  "title": "Субботник",
  "description": "...",
  "location": "Парк",
  "image": "/static/images/events/subbotnik.jpg",
  "start_date": "2026-07-01T10:00:00Z",
  "end_date": "2026-07-01T14:00:00Z",
  "registration_deadline": "2026-06-30T23:59:00Z",
  "max_participants": 50,
  "reserve_participants": 10,
  "skill_points": 50,
  "template_id": 1
}

→ 201 Created
```

## 🎨 UI-флоу для фронта

### Юзер (Volunteer):
1. Логин → главная страница со списком ивентов
2. Клик на ивент → детали + кнопка "Зарегистрироваться"
3. Профиль → история ивентов, баланс SP, каталог наград
4. Battle Pass → список наград с кнопкой "Купить" (если `available: true`)
5. После покупки → награда в "Мои награды" со статусом "Ожидает выдачи"

### Организатор:
1. Всё что юзер + создание ивентов
2. После ивента → список участников + кнопка "Подтвердить" у каждого
3. Чёрный список → добавление/удаление юзеров

### Админ:
1. Всё что орг
2. Поиск юзеров → карточка с 3 кнопками:
   - 🔼 Повысить роль
   - 💰 Начислить SP
   - 🎁 Выдать мерч (показывает купленные награды с `picked_up: false`)
3. Подтверждение ивентов → массовое начисление SP
4. Управление наградами и шаблонами

## 🚀 Запуск

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

Сервер на `http://localhost:8080`.

## 📋 Коды ошибок

| Код | HTTP | Описание |
|-----|------|----------|
| `INVALID_JSON` | 400 | Неверный JSON |
| `MISSING_FIELDS` | 400 | Обязательные поля отсутствуют |
| `USER_EXISTS` | 409 | Логин уже занят |
| `INVALID_CREDENTIALS` | 401 | Неверный логин/пароль |
| `UNAUTHORIZED` | 401 | Нет токена |
| `FORBIDDEN` | 403 | Недостаточно прав |
| `EVENT_NOT_FOUND` | 404 | Ивент не найден |
| `USER_NOT_FOUND` | 404 | Юзер не найден |
| `REGISTRATION_CLOSED` | 400 | Регистрация закрыта |
| `INSUFFICIENT_POINTS` | 400 | Не хватает SP |
| `ALREADY_CLAIMED` | 400 | Награда уже получена |
