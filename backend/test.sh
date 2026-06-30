#!/bin/bash
# Тестовый скрипт для Volunteer Platform API

set -o pipefail

BASE_URL="http://localhost:8080"

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

# Уникальные логины
TIMESTAMP=$(date +%s)
VOLUNTEER_LOGIN="volunteer_${TIMESTAMP}"
ORGANIZER_LOGIN="organizer_${TIMESTAMP}"

# Извлечение токена из ответа
extract_token() {
    echo "$1" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('token', ''))" 2>/dev/null
}

# Извлечение ID из ответа
extract_id() {
    echo "$1" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('id', 0))" 2>/dev/null
}

print_header() {
    echo -e "\n${YELLOW}=== $1 ===${NC}" >&2
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}" >&2
    ((PASSED++))
}

print_error() {
    echo -e "${RED}✗ $1${NC}" >&2
    ((FAILED++))
}

# Выполнение запроса
do_request() {
    local method=$1
    local url=$2
    local data=$3
    local token=$4
    local expected_status=$5
    local description=$6

    local curl_args=(-s -w "\n%{http_code}" -X "$method" "${BASE_URL}${url}" -H "Content-Type: application/json")

    if [ -n "$token" ]; then
        curl_args+=(-H "Authorization: Bearer $token")
    fi

    if [ -n "$data" ]; then
        curl_args+=(-d "$data")
    fi

    local response
    response=$(curl "${curl_args[@]}")

    local http_code
    http_code=$(echo "$response" | tail -n1)
    local body
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq "$expected_status" ]; then
        print_success "$description (HTTP $http_code)"
    else
        print_error "$description (ожидается $expected_status, получено $http_code)"
        echo "  Response: $body" >&2
    fi

    echo "$body"
}

# Проверка токена
check_token() {
    local token=$1
    local name=$2
    if [ -z "$token" ]; then
        print_error "Не удалось получить токен для $name"
        echo -e "${RED}Прерывание тестов${NC}" >&2
        exit 1
    fi
}

# ==================== ТЕСТЫ ====================

print_header "0. Проверка сервера"
HEALTH=$(do_request "GET" "/health" "" "" 200 "Health check")

print_header "1. Регистрация пользователей"
VOLUNTEER_RESPONSE=$(do_request "POST" "/auth/register" \
    "{\"login\":\"${VOLUNTEER_LOGIN}\",\"password\":\"pass123\",\"first_name\":\"Иван\",\"last_name\":\"Иванов\",\"middle_name\":\"Иванович\",\"telegram\":\"@ivanov\"}" \
    "" 201 "Регистрация волонтёра (${VOLUNTEER_LOGIN})")
VOLUNTEER_TOKEN=$(extract_token "$VOLUNTEER_RESPONSE")
check_token "$VOLUNTEER_TOKEN" "волонтёра"

ORG_RESPONSE=$(do_request "POST" "/auth/register" \
    "{\"login\":\"${ORGANIZER_LOGIN}\",\"password\":\"pass123\",\"first_name\":\"Пётр\",\"last_name\":\"Петров\",\"middle_name\":\"\",\"telegram\":\"@petrov\"}" \
    "" 201 "Регистрация организатора (${ORGANIZER_LOGIN})")
ORG_TOKEN=$(extract_token "$ORG_RESPONSE")
check_token "$ORG_TOKEN" "организатора"

print_header "2. Логин"
LOGIN_RESPONSE=$(do_request "POST" "/auth/login" \
    "{\"login\":\"${VOLUNTEER_LOGIN}\",\"password\":\"pass123\"}" \
    "" 200 "Логин волонтёра")

do_request "POST" "/auth/login" \
    "{\"login\":\"${VOLUNTEER_LOGIN}\",\"password\":\"wrongpass\"}" \
    "" 401 "Неверный пароль (должен быть 401)"

print_header "3. Профиль"
do_request "GET" "/auth/me" "" "$VOLUNTEER_TOKEN" 200 "Получение профиля волонтёра"
do_request "GET" "/auth/me" "" "" 401 "Профиль без токена (должен быть 401)"

print_header "4. Создание ивента (нужен токен админа)"
ADMIN_LOGIN=$(grep ADMIN_LOGIN .env 2>/dev/null | cut -d'=' -f2 | tr -d ' ')
ADMIN_PASS=$(grep ADMIN_PASSWORD .env 2>/dev/null | cut -d'=' -f2 | tr -d ' ')

if [ -z "$ADMIN_LOGIN" ] || [ -z "$ADMIN_PASS" ]; then
    print_error "Не удалось прочитать ADMIN_LOGIN/ADMIN_PASSWORD из .env"
    exit 1
fi

ADMIN_RESPONSE=$(do_request "POST" "/auth/login" \
    "{\"login\":\"${ADMIN_LOGIN}\",\"password\":\"${ADMIN_PASS}\"}" \
    "" 200 "Логин админа")
ADMIN_TOKEN=$(extract_token "$ADMIN_RESPONSE")
check_token "$ADMIN_TOKEN" "админа"

EVENT_RESPONSE=$(do_request "POST" "/events" \
    "{\"title\":\"Субботник\",\"description\":\"Уборка парка\",\"location\":\"Парк\",\"image\":\"/static/images/events/subbotnik.jpg\",\"start_date\":\"2026-07-01T10:00:00Z\",\"end_date\":\"2026-07-01T14:00:00Z\",\"registration_deadline\":\"2026-06-30T23:59:00Z\",\"max_participants\":50,\"reserve_participants\":10,\"skill_points\":50}" \
    "$ADMIN_TOKEN" 201 "Создание ивента")
EVENT_ID=$(extract_id "$EVENT_RESPONSE")

if [ -z "$EVENT_ID" ] || [ "$EVENT_ID" = "0" ]; then
    print_error "Не удалось получить EVENT_ID"
    exit 1
fi
echo "  Event ID: $EVENT_ID" >&2

print_header "5. Список ивентов"
do_request "GET" "/events" "" "$VOLUNTEER_TOKEN" 200 "Получение списка ивентов"
do_request "GET" "/events/$EVENT_ID" "" "$VOLUNTEER_TOKEN" 200 "Получение ивента по ID"

print_header "6. Регистрация на ивент"
do_request "POST" "/events/$EVENT_ID/register" "" "$VOLUNTEER_TOKEN" 200 "Регистрация волонтёра на ивент"
do_request "POST" "/events/$EVENT_ID/register" "" "$VOLUNTEER_TOKEN" 400 "Повторная регистрация (должна быть ошибка)"

print_header "7. Участники ивента"
do_request "GET" "/events/$EVENT_ID/participants" "" "$ADMIN_TOKEN" 200 "Список участников"

print_header "8. Отмена регистрации"
do_request "DELETE" "/events/$EVENT_ID/register" "" "$VOLUNTEER_TOKEN" 200 "Отмена регистрации"
do_request "DELETE" "/events/$EVENT_ID/register" "" "$VOLUNTEER_TOKEN" 400 "Повторная отмена (должна быть ошибка)"

print_header "9. История ивентов"
do_request "GET" "/users/me/events" "" "$VOLUNTEER_TOKEN" 200 "История ивентов волонтёра"

print_header "10. Админские действия"
# Получаем ID волонтёра
ME_RESPONSE=$(do_request "GET" "/auth/me" "" "$VOLUNTEER_TOKEN" 200 "Получение ID волонтёра")
VOLUNTEER_ID=$(echo "$ME_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('id', 0))" 2>/dev/null)
echo "  Volunteer ID: $VOLUNTEER_ID" >&2

# Получаем ID орга
ORG_ME_RESPONSE=$(do_request "GET" "/auth/me" "" "$ORG_TOKEN" 200 "Получение ID орга")
ORGANIZER_ID=$(echo "$ORG_ME_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('data', {}).get('id', 0))" 2>/dev/null)
echo "  Organizer ID: $ORGANIZER_ID" >&2

# Повышаем волонтёра до Organizer
do_request "POST" "/admin/users/${VOLUNTEER_ID}/promote" \
    '{"role":"Organizer"}' \
    "$ADMIN_TOKEN" 200 "Повышение волонтёра до Organizer"

# Повышаем орга до Organizer
do_request "POST" "/admin/users/${ORGANIZER_ID}/promote" \
    '{"role":"Organizer"}' \
    "$ADMIN_TOKEN" 200 "Повышение орга до Organizer"

# Начисление SP
do_request "POST" "/admin/users/${VOLUNTEER_ID}/award" \
    '{"points":100,"reason":"Тестовое начисление"}' \
    "$ADMIN_TOKEN" 200 "Начисление SP вручную"

# Поиск юзеров
do_request "GET" "/admin/users?first_name=Иван&last_name=Иванов" "" "$ADMIN_TOKEN" 200 "Поиск юзеров"

print_header "11. Награды (Battle Pass)"
REWARD_RESPONSE=$(do_request "POST" "/admin/rewards" \
    '{"name":"Худи","description":"Лимитированная коллекция","cost":50,"image_url":"/static/images/rewards/hoodie.png"}' \
    "$ADMIN_TOKEN" 201 "Создание награды")
REWARD_ID=$(extract_id "$REWARD_RESPONSE")

do_request "GET" "/admin/rewards" "" "$ADMIN_TOKEN" 200 "Список всех наград"

# Начислим ещё SP чтобы хватило
do_request "POST" "/admin/users/${VOLUNTEER_ID}/award" \
    '{"points":100,"reason":"Для покупки"}' \
    "$ADMIN_TOKEN" 200 "Начисление SP для покупки"

do_request "GET" "/users/me/rewards" "" "$VOLUNTEER_TOKEN" 200 "Мои награды (с флагами)"

do_request "POST" "/users/me/rewards/$REWARD_ID/claim" "" "$VOLUNTEER_TOKEN" 200 "Покупка награды"
do_request "POST" "/users/me/rewards/$REWARD_ID/claim" "" "$VOLUNTEER_TOKEN" 400 "Повторная покупка (должна быть ошибка)"

print_header "12. Выдача мерча (админ)"
do_request "GET" "/admin/users/${VOLUNTEER_ID}/rewards" "" "$ADMIN_TOKEN" 200 "Купленные награды юзера"
do_request "POST" "/admin/users/${VOLUNTEER_ID}/rewards/$REWARD_ID/pickup" "" "$ADMIN_TOKEN" 200 "Отметка о выдаче"

print_header "13. История транзакций SP"
do_request "GET" "/users/${VOLUNTEER_ID}/skill-points/history" "" "$VOLUNTEER_TOKEN" 200 "История транзакций"

print_header "14. Шаблоны ивентов"
do_request "POST" "/organizer/event-templates" \
    '{"title":"Лекция","description":"Обычная лекция","location":"Аудитория 101","image_url":"/static/images/events/lecture.jpg","duration_minutes":90,"max_participants":100,"reserve_participants":20,"skill_points":30}' \
    "$ORG_TOKEN" 201 "Создание шаблона"
do_request "GET" "/organizer/event-templates" "" "$ORG_TOKEN" 200 "Список шаблонов"

print_header "15. Картинки"
do_request "GET" "/assets/images" "" "$VOLUNTEER_TOKEN" 200 "Список доступных картинок"

print_header "16. Бан-лист (нужен Organizer)"
# Логинимся как org (теперь он Organizer)
ORG_LOGIN_RESPONSE=$(do_request "POST" "/auth/login" \
    "{\"login\":\"${ORGANIZER_LOGIN}\",\"password\":\"pass123\"}" \
    "" 200 "Логин орга")
ORG_TOKEN=$(extract_token "$ORG_LOGIN_RESPONSE")

do_request "POST" "/organizer/blacklist" \
    "{\"user_id\":${VOLUNTEER_ID}}" \
    "$ORG_TOKEN" 200 "Бан юзера"
do_request "DELETE" "/organizer/blacklist/${VOLUNTEER_ID}" "" "$ORG_TOKEN" 200 "Разбан юзера"

# ==================== ИТОГИ ====================
echo -e "\n${YELLOW}=== РЕЗУЛЬТАТЫ ===${NC}" >&2
echo -e "${GREEN}Пройдено: $PASSED${NC}" >&2
echo -e "${RED}Провалено: $FAILED${NC}" >&2

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}Все тесты пройдены! 🎉${NC}" >&2
    exit 0
else
    echo -e "${RED}Есть проваленные тесты${NC}" >&2
    exit 1
fi