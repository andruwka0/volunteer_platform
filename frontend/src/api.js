/**
 * @typedef {Object} User
 * @property {number} ID
 * @property {string} Login
 * @property {string} FirstName
 * @property {string} LastName
 * @property {string} Telegram
 * @property {number} SkillPoints
 * @property {string} Role
 */

/**
 * @typedef {Object} Event
 * @property {number} ID
 * @property {string} Title
 * @property {string} Description
 * @property {string} Location
 * @property {string} CoverImageURL
 * @property {string} Status
 * @property {string} StartDate
 * @property {string} EndDate
 * @property {string} [RegistrationDeadline] - Дедлайн регистрации (опционально)
 * @property {number} SkillPoints
 * @property {number} ParticipantsCount
 * @property {number} CreatedByID
 * @property {number} [MaxParticipants] - Максимум участников (опционально)
 * @property {number} [ReserveParticipants] - Резерв участников
 * @property {number} [ReserveCount] - Количество в резерве
 */

/**
 * @typedef {Object} TokenResponse
 * @property {string} token
 */

/**
 * @typedef {Object} ErrorResponse
 * @property {string} [message]
 */

/**
 * @typedef {Object} ApiOptions
 * @property {string} [method] - HTTP метод (GET, POST, DELETE и т.д.)
 * @property {Object|FormData|string|null} [body] - Тело запроса (объект, FormData или строка)
 * @property {Object} [headers] - Заголовки запроса
 */

/**
 * Делает запрос к API
 * @param {string} endpoint - URL эндпоинта
 * @param {ApiOptions} [options] - Опции запроса
 * @returns {Promise<any>}
 */
async function api(endpoint, options = {}) {
    const token = localStorage.getItem('token');
    const headers = {};
    if (options.headers) Object.assign(headers, options.headers);
    if (token) headers['Authorization'] = token;

    let body = options.body;
    const isFormData = body instanceof FormData;

    // Если body — объект (не FormData), сериализуем в JSON
    if (!isFormData && body && typeof body === 'object') {
        headers['Content-Type'] = 'application/json';
        body = JSON.stringify(body);
    }

    /** @type {RequestInit} */
    const fetchOptions = { method: options.method || 'GET', headers };
    if (body) fetchOptions.body = body;

    const res = await fetch(endpoint, fetchOptions);
    /** @type {ErrorResponse} */
    const data = await res.json().catch(() => ({}));

    if (!res.ok) throw new Error(data.message || `Ошибка ${res.status}`);
    return data;
}

function redirect(path) { window.location.href = path; }
function getParam(name) { return new URLSearchParams(window.location.search).get(name); }

/** @returns {Promise<User|null>} */
async function requireAuth() {
    if (!localStorage.getItem('token')) { redirect('/static/login.html'); return null; }
    try {
        return await api('/auth/me');
    } catch {
        localStorage.removeItem('token');
        redirect('/static/login.html');
        return null;
    }
}

function showFormError(formId, message) {
    const form = document.getElementById(formId);
    let err = form.querySelector('.alert-error');
    if (!err) { err = document.createElement('div'); err.className = 'alert alert-error'; form.prepend(err); }
    err.textContent = message;
    err.style.display = 'block';
}

/**
 * Конвертирует значение из input type="datetime-local" в RFC3339
 * "2026-06-12T18:15" → "2026-06-12T18:15:00+03:00"
 * @param {string} datetimeLocal
 * @returns {string|null}
 */
function toRFC3339(datetimeLocal) {
    if (!datetimeLocal) return null;
    const d = new Date(datetimeLocal);
    const offsetMin = -d.getTimezoneOffset();
    const sign = offsetMin >= 0 ? '+' : '-';
    const absMin = Math.abs(offsetMin);
    const h = String(Math.floor(absMin / 60)).padStart(2, '0');
    const m = String(absMin % 60).padStart(2, '0');
    return `${datetimeLocal}:00${sign}${h}:${m}`;
}