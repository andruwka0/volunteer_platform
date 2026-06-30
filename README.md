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

**JWT токены** в заголовке `Authorization` с префиксом `Bearer `:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Токен получается после `POST /auth/login` или `POST /auth/register` в поле `data.token`.

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
| **Volunteer** | Регистрироваться на ивенты, покупать награды, смотреть свой профиль, историю ивентов и SP |
| **Organizer** | + Создавать ивенты, подтверждать посещаемость СВОИХ ивентов, банить юзеров |
| **Admin** | + Всё: управление ролями, начисление SP, подтверждение ивентов, создание наград и шаблонов, выдача мерча |

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

### Защищённые (нужен Auth — любой авторизованный юзер)

#### Мой профиль
```http
GET /auth/me
Authorization: Bearer <token>

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
Authorization: Bearer <token>

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

⚠️ Забаненные юзеры **не видят** ивенты от конкретного организатора в этом списке.

#### Детали ивента
```http
GET /events/{id}
Authorization: Bearer <token>

→ 200 OK
```

#### Регистрация на ивент
```http
POST /events/{id}/register
Authorization: Bearer <token>

→ 200 OK
{
  "success": true,
  "data": { "message": "Вы успешно зарегистрированы на мероприятие" }
}
```

⚠️ Если мест нет — юзер автоматически попадает в **резерв**.

#### Отмена регистрации
```http
DELETE /events/{id}/register
Authorization: Bearer <token>

→ 200 OK
```

⚠️ Если юзер был в основном списке — первый из резерва автоматически повышается.

#### Участники ивента
```http
GET /events/{id}/participants
Authorization: Bearer <token>

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
Authorization: Bearer <token>

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
Authorization: Bearer <token>

→ 200 OK
{
  "success": true,
  "data": {
    "user_id": 5,
    "transactions": [
      {
        "ID": 1,
        "UserID": 5,
        "Points": 50,
        "Type": "event",
        "Reason": "Участие в мероприятии: Субботник",
        "EventID": 1,
        "CreatedAt": "2026-07-01T15:00:00Z"
      }
    ],
    "count": 1
  }
}
```

⚠️ **Важно:** юзер может смотреть только **свою** историю. Админ — любую.  
⚠️ Поля возвращаются в **CamelCase** (как в Go), т.к. `SkillPointTransaction` сериализуется напрямую без DTO.

**Типы транзакций:**
- `manual` — ручное начисление админом
- `event` — автоматическое за ивент
- `reward` — списание за награду

### Создание ивента (Organizer / Admin)

```http
POST /events
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Субботник",
  "description": "Уборка парка",
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

⚠️ **Требует роль Organizer или Admin.** Обычный Volunteer получит `403 FORBIDDEN`.  
💡 `template_id` — опционально, ссылка на шаблон.

### Каталог наград (Battle Pass)

#### Все награды (каталог)
```http
GET /admin/rewards
Authorization: Bearer <token>

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
Authorization: Bearer <token>

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
Authorization: Bearer <token>

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
Authorization: Bearer <token>

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
Authorization: Bearer <token>

→ 200 OK
{
  "success": true,
  "data": { "message": "Посещаемость подтверждена" }
}
```

⚠️ Только **создатель ивента** может подтверждать!

#### Забанить юзера
```http
POST /organizer/blacklist
Authorization: Bearer <token>
Content-Type: application/json

{ "user_id": 5 }

→ 200 OK
```

Забаненный юзер **не видит** ивенты этого орга в списке.

#### Разбанить
```http
DELETE /organizer/blacklist/{user_id}
Authorization: Bearer <token>

→ 200 OK
```

### Админские роуты (только Admin)

#### Поиск юзеров
```http
GET /admin/users?first_name=Иван&last_name=Иванов
Authorization: Bearer <token>

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
Authorization: Bearer <token>

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
Authorization: Bearer <token>

→ 200 OK
```

#### Повысить роль
```http
POST /admin/users/{id}/promote
Authorization: Bearer <token>
Content-Type: application/json

{ "role": "Organizer" }

→ 200 OK
```

Возможные роли: `Organizer`, `Admin`.

#### Начислить SP вручную
```http
POST /admin/users/{id}/award
Authorization: Bearer <token>
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
Authorization: Bearer <token>

→ 200 OK
```

⚠️ Работает только для ивентов в статусе `EVENT-FINISHED`. SP начисляются только тем, у кого `attendance_confirmed: true`.

#### Создать награду
```http
POST /admin/rewards
Authorization: Bearer <token>
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

### Оргские роуты (Organizer / Admin)

#### Создать шаблон ивента
```http
POST /organizer/event-templates
Authorization: Bearer <token>
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
GET /organizer/event-templates
Authorization: Bearer <token>

→ 200 OK
```

## 🎨 UI-флоу для фронта

### Юзер (Volunteer):
1. **Логин** → главная страница со списком ивентов
2. **Клик на ивент** → детали + кнопка "Зарегистрироваться"
3. **Профиль** → история ивентов, баланс SP, каталог наград
4. **Battle Pass** → список наград с кнопкой "Купить" (если `available: true`)
5. **После покупки** → награда в "Мои награды" со статусом "Ожидает выдачи"

### Организатор:
1. Всё что юзер + **создание ивентов**
2. **После ивента** → список участников + кнопка "Подтвердить" у каждого
3. **Чёрный список** → добавление/удаление юзеров

### Админ:
1. Всё что орг
2. **Поиск юзеров** → карточка с 3 кнопками:
   - 🔼 Повысить роль
   - 💰 Начислить SP
   - 🎁 Выдать мерч (показывает купленные награды с `picked_up: false`)
3. **Подтверждение ивентов** → массовое начисление SP
4. **Управление наградами и шаблонами**

## 🚀 Запуск

```bash
cd backend

# Создай .env из шаблона
cp .env.example .env
# Отредактируй ADMIN_LOGIN, ADMIN_PASSWORD, JWT_SECRET

# Установи зависимости
go mod tidy

# Запусти
go run cmd/server/main.go
```

Сервер на `http://localhost:8080`.

## 🔧 Переменные окружения (backend/.env)

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `ADMIN_LOGIN` | Логин админа (создаётся при старте) | — |
| `ADMIN_PASSWORD` | Пароль админа | — |
| `JWT_SECRET` | Секрет для подписи JWT (мин. 32 символа) | `default-dev-secret-...` |
| `JWT_EXPIRATION_HOURS` | Срок жизни токена в часах | `24` |

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
| `NOT_CREATOR` | 403 | Только создатель ивента может подтверждать |
| `ALREADY_CONFIRMED` | 400 | Посещаемость уже подтверждена |
| `EVENT_NOT_FINISHED` | 400 | Ивент ещё не закончился |
| `ORGANIZER_SELF_REG` | 400 | Организатор не может регаться на свои ивенты |

## ⚠️ Важные замечания для фронтендера

1. **Токен с `Bearer `** — `Authorization: Bearer <token>` (стандарт HTTP)
2. **Создание ивентов** — `POST /events` требует роль **Organizer** или **Admin** (Volunteer получит 403)
3. **Шаблоны** —  для **Admin** или **Organizer** (`/organizer/event-templates`)
4. **CamelCase в истории SP** — `/users/{id}/skill-points/history` возвращает поля в CamelCase (как в Go), остальные эндпоинты — в snake_case
5. **Автоматическая смена статусов** — воркер меняет статусы ивентов по датам (RECRUITING → ACTIVE → FINISHED)
6. **Резерв** — если мест нет, юзер автоматически попадает в резерв; при отмене — первый из резерва повышается

## 📸 Картинки

Картинки лежат в `backend/app/static/images/`:
- `rewards/` — картинки наград
- `events/` — картинки ивентов

При старте бэкенд автоматически сканирует эту папку и делает картинки доступными через `GET /assets/images`.
