const TOKEN_KEY = 'volunteer_platform_token';
const state = {
  token: localStorage.getItem(TOKEN_KEY) || '',
  user: null,
  events: [],
  participantsByEvent: new Map(),
};

const selectors = {
  authSection: document.querySelector('#auth-section'),
  sessionLabel: document.querySelector('#session-label'),
  logoutButton: document.querySelector('#logout-button'),
  profile: document.querySelector('#profile'),
  profileBody: document.querySelector('#profile-body'),
  create: document.querySelector('#create'),
  admin: document.querySelector('#admin'),
  eventsList: document.querySelector('#events-list'),
  toastRoot: document.querySelector('#toast-root'),
};

function toast(message, category = 'info') {
  const element = document.createElement('div');
  element.className = `toast ${category === 'error' ? 'toast-error' : 'toast-info'}`;
  element.textContent = message;
  selectors.toastRoot.append(element);
  setTimeout(() => element.remove(), 4500);
}


function escapeHtml(value) {
  return String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

function readField(object, ...keys) {
  for (const key of keys) {
    if (object && object[key] !== undefined && object[key] !== null) return object[key];
  }
  return undefined;
}

function eventId(event) {
  return readField(event, 'ID', 'id');
}

function userId(user) {
  return readField(user, 'ID', 'id');
}

function roleOf(user) {
  return readField(user, 'Role', 'role') || 'Guest';
}

function canManageEvents() {
  return ['Admin', 'Organizer'].includes(roleOf(state.user));
}

function isAdmin() {
  return roleOf(state.user) === 'Admin';
}

async function api(path, options = {}) {
  const headers = new Headers(options.headers || {});
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  if (state.token) {
    headers.set('Authorization', state.token);
  }

  const response = await fetch(path, { ...options, headers });
  const contentType = response.headers.get('Content-Type') || '';
  const payload = contentType.includes('application/json') ? await response.json() : await response.text();

  if (!response.ok) {
    const message = typeof payload === 'object' ? payload.message : payload;
    throw new Error(message || `HTTP ${response.status}`);
  }

  return payload;
}

function formDataToObject(form) {
  return Object.fromEntries(new FormData(form).entries());
}

function toIsoOrNull(value) {
  return value ? new Date(value).toISOString() : null;
}

function numberOrNull(value) {
  if (value === '' || value === undefined || value === null) return null;
  const number = Number(value);
  return Number.isFinite(number) ? number : null;
}

function formatDate(value) {
  if (!value) return '—';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString('ru-RU');
}

function statusLabel(status) {
  const labels = {
    'EVENT-RECRUITING': 'идёт набор',
    'EVENT-ACTIVE': 'активно',
    'EVENT-FINISHED': 'завершено',
    'EVENT-CLOSED': 'закрыто',
    'EVENT-CANCELLED': 'отменено',
  };
  return labels[status] || status || 'без статуса';
}

function renderProfile() {
  const user = state.user;
  selectors.authSection.hidden = Boolean(user);
  selectors.logoutButton.hidden = !user;
  selectors.profile.hidden = !user;
  selectors.create.hidden = !canManageEvents();
  selectors.admin.hidden = !isAdmin();

  if (!user) {
    selectors.sessionLabel.textContent = 'Гость';
    selectors.profileBody.innerHTML = '';
    return;
  }

  const name = [readField(user, 'FirstName', 'first_name'), readField(user, 'LastName', 'last_name')]
    .filter(Boolean)
    .join(' ') || readField(user, 'Login', 'login') || `Пользователь #${userId(user)}`;
  selectors.sessionLabel.textContent = `${name} · ${roleOf(user)}`;
  selectors.profileBody.innerHTML = `
    <div><span class="text-muted">ID</span><b>${escapeHtml(userId(user))}</b></div>
    <div><span class="text-muted">Логин</span><b>${escapeHtml(readField(user, 'Login', 'login') || '—')}</b></div>
    <div><span class="text-muted">Имя</span><b>${escapeHtml(name)}</b></div>
    <div><span class="text-muted">Telegram</span><b>${escapeHtml(readField(user, 'Telegram', 'telegram') || '—')}</b></div>
    <div><span class="text-muted">Роль</span><b>${escapeHtml(roleOf(user))}</b></div>
    <div><span class="text-muted">Skill Points</span><b>${escapeHtml(readField(user, 'SkillPoints', 'skill_points') || 0)}</b></div>
  `;
}

async function loadMe() {
  if (!state.token) {
    state.user = null;
    renderProfile();
    return;
  }

  try {
    state.user = await api('/auth/me');
  } catch (error) {
    localStorage.removeItem(TOKEN_KEY);
    state.token = '';
    state.user = null;
    toast(`Сессия сброшена: ${error.message}`, 'error');
  }
  renderProfile();
}

async function loadParticipants(id) {
  try {
    const data = await api(`/events/${id}/participants`);
    state.participantsByEvent.set(Number(id), data.participants || []);
  } catch {
    state.participantsByEvent.set(Number(id), []);
  }
}

async function loadEvents() {
  selectors.eventsList.innerHTML = '<article class="card"><p class="text-muted">Загрузка мероприятий…</p></article>';
  try {
    state.events = await api('/events');
    await Promise.all(state.events.map((event) => loadParticipants(eventId(event))));
    renderEvents();
  } catch (error) {
    selectors.eventsList.innerHTML = `<article class="card"><p class="text-muted">${escapeHtml(error.message)}</p></article>`;
  }
}

function renderEvents() {
  if (!state.events.length) {
    selectors.eventsList.innerHTML = '<article class="card"><p class="text-muted">Пока нет мероприятий.</p></article>';
    return;
  }

  selectors.eventsList.innerHTML = state.events.map((event) => {
    const id = eventId(event);
    const image = readField(event, 'CoverImageURL', 'cover_image_url', 'image') || '/static/img/placeholder.svg';
    const max = readField(event, 'MaxParticipants', 'max_participants');
    const participants = state.participantsByEvent.get(Number(id)) || [];
    const isRegistered = state.user && participants.includes(userId(state.user));
    const status = readField(event, 'Status', 'status');
    const canRegister = state.user && status === 'EVENT-RECRUITING';
    return `
      <article class="card event-row">
        <img class="event-avatar" src="${escapeHtml(image)}" alt="">
        <div class="event-main">
          <div class="event-head">
            <div>
              <h3 class="event-title">#${escapeHtml(id)} · ${escapeHtml(readField(event, 'Title', 'title') || 'Без названия')}</h3>
              <p class="text-muted">${escapeHtml(readField(event, 'Description', 'description') || 'Описание не указано')}</p>
            </div>
            <div class="event-head-chips">
              <span class="badge">${escapeHtml(statusLabel(status))}</span>
              <span class="badge badge-success">${escapeHtml(readField(event, 'SkillPoints', 'skill_points') || 0)} SP</span>
            </div>
          </div>
          <div class="event-meta">
            <span>📍 ${escapeHtml(readField(event, 'Location', 'location') || '—')}</span>
            <span>🕒 ${escapeHtml(formatDate(readField(event, 'StartDate', 'start_date')))} — ${escapeHtml(formatDate(readField(event, 'EndDate', 'end_date')))}</span>
            <span>⏳ регистрация до ${escapeHtml(formatDate(readField(event, 'RegistrationDeadline', 'registration_deadline')))}</span>
            <span>👥 ${escapeHtml(readField(event, 'ParticipantsCount', 'participants_count') || 0)}${max ? ` / ${escapeHtml(max)}` : ''}</span>
            <span>🧾 участники: ${escapeHtml(participants.length ? participants.join(', ') : 'нет')}</span>
            <span>👤 организатор ID: ${escapeHtml(readField(event, 'CreatedByID', 'created_by_id') || '—')}</span>
          </div>
          <div class="action-row">
            <button class="btn btn-primary" type="button" data-register-event="${escapeHtml(id)}" ${!canRegister || isRegistered ? 'disabled' : ''}>${escapeHtml(isRegistered ? 'Вы зарегистрированы' : 'Зарегистрироваться')}</button>
            <button class="btn btn-secondary" type="button" data-cancel-event="${escapeHtml(id)}" ${!isRegistered ? 'disabled' : ''}>Отменить регистрацию</button>
          </div>
        </div>
      </article>
    `;
  }).join('');
}

async function authenticate(path, form) {
  const data = formDataToObject(form);
  const result = await api(path, {
    method: 'POST',
    body: JSON.stringify(data),
  });
  state.token = result.token;
  localStorage.setItem(TOKEN_KEY, state.token);
  form.reset();
  await loadMe();
  await loadEvents();
}

function bindForms() {
  document.querySelector('#login-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    try {
      await authenticate('/auth/login', event.currentTarget);
      toast('Вход выполнен');
    } catch (error) {
      toast(error.message, 'error');
    }
  });

  document.querySelector('#register-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    try {
      await authenticate('/auth/register', event.currentTarget);
      toast('Аккаунт создан');
    } catch (error) {
      toast(error.message, 'error');
    }
  });

  document.querySelector('#event-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formDataToObject(form);
    const payload = {
      title: data.title,
      description: data.description,
      location: data.location,
      image: data.image,
      start_date: toIsoOrNull(data.start_date),
      end_date: toIsoOrNull(data.end_date),
      registration_deadline: toIsoOrNull(data.registration_deadline),
      max_participants: numberOrNull(data.max_participants),
      reserve_participants: numberOrNull(data.reserve_participants) || 0,
      skill_points: numberOrNull(data.skill_points) || 0,
    };

    try {
      await api('/events', { method: 'POST', body: JSON.stringify(payload) });
      form.reset();
      toast('Мероприятие создано');
      await loadEvents();
    } catch (error) {
      toast(error.message, 'error');
    }
  });

  document.querySelector('#approve-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    const { event_id: id } = formDataToObject(event.currentTarget);
    try {
      const result = await api(`/admin/events/${id}/approve`, { method: 'POST' });
      toast(result.message || 'Мероприятие закрыто');
      await loadEvents();
      await loadMe();
    } catch (error) {
      toast(error.message, 'error');
    }
  });

  document.querySelector('#promote-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    const { user_id: id, role } = formDataToObject(event.currentTarget);
    try {
      const result = await api(`/admin/users/${id}/promote`, {
        method: 'POST',
        body: JSON.stringify({ role }),
      });
      toast(result.message || 'Роль обновлена');
    } catch (error) {
      toast(error.message, 'error');
    }
  });
}

function bindActions() {
  selectors.logoutButton?.addEventListener('click', async () => {
    localStorage.removeItem(TOKEN_KEY);
    state.token = '';
    state.user = null;
    renderProfile();
    renderEvents();
    toast('Вы вышли из аккаунта');
  });

  document.querySelector('#refresh-events')?.addEventListener('click', loadEvents);

  document.addEventListener('click', async (event) => {
    const registerButton = event.target.closest('[data-register-event]');
    const cancelButton = event.target.closest('[data-cancel-event]');
    const id = registerButton?.dataset.registerEvent || cancelButton?.dataset.cancelEvent;
    if (!id) return;

    try {
      const result = await api(`/events/${id}/register`, {
        method: registerButton ? 'POST' : 'DELETE',
      });
      toast(result.message || 'Готово');
      await loadEvents();
    } catch (error) {
      toast(error.message, 'error');
    }
  });
}

bindForms();
bindActions();
await loadMe();
await loadEvents();
