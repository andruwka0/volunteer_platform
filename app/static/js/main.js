document.querySelectorAll('.card').forEach((el, idx) => {
  el.classList.add('reveal');
  el.style.animationDelay = `${Math.min(idx * 60, 360)}ms`;
});

document.querySelectorAll('[data-deadline]').forEach((card) => {
  const timerEl = card.querySelector('.event-timer');
  if (!timerEl) return;
  const raw = card.getAttribute('data-deadline');
  if (!raw) { timerEl.textContent = 'дата окончания не задана'; return; }
  const end = new Date(raw);
  const tick = () => {
    const diff = end - new Date();
    if (diff <= 0) { timerEl.textContent = 'регистрация закрыта'; return; }
    const totalMin = Math.floor(diff / 60000);
    const days = Math.floor(totalMin / (60*24));
    const h = Math.floor((totalMin % (60*24)) / 60);
    const m = totalMin % 60;
    timerEl.textContent = `до конца регистрации: ${days}д ${h}ч ${String(m).padStart(2, '0')}м`;
  };
  tick();
  setInterval(tick, 60000);
});

document.querySelectorAll('[data-multi-select]').forEach((block) => {
  const btn = block.querySelector('[data-multi-select-toggle]');
  const inputs = Array.from(block.querySelectorAll('[data-multi-select-input]'));
  const search = block.querySelector('[data-multi-select-search]');
  const renderLabel = () => {
    const selected = inputs.filter((x) => x.checked);
    inputs.forEach((x) => x.closest('label')?.classList.toggle('selected', x.checked));
    if (!selected.length) {
      btn.textContent = 'выбрать участников';
      return;
    }
    const first = selected[0].closest('label')?.querySelector('span')?.textContent?.trim() || 'участник';
    btn.textContent = selected.length === 1 ? first : `${first} +${selected.length - 1}`;
  };
  btn?.addEventListener('click', () => block.classList.toggle('open'));
  inputs.forEach((i) => i.addEventListener('change', renderLabel));
  search?.addEventListener('input', () => {
    const q = search.value.trim().toLowerCase();
    block.classList.add('open');
    inputs.forEach((i) => {
      const label = i.closest('label');
      if (!label) return;
      const text = label.textContent.toLowerCase();
      label.style.display = text.includes(q) ? 'flex' : 'none';
    });
  });
  renderLabel();
});

document.addEventListener('click', (e) => {
  document.querySelectorAll('[data-multi-select].open').forEach((block) => {
    if (!block.contains(e.target)) block.classList.remove('open');
  });
});

document.querySelectorAll('#toast-root .toast').forEach((toast) => {
  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateY(8px)';
    setTimeout(() => toast.remove(), 220);
  }, 2600);
});

const bindImagePreview = (inputSelector, previewSelector) => {
  const input = document.querySelector(inputSelector);
  const preview = document.querySelector(previewSelector);
  if (!input || !preview) return;
  input.addEventListener('change', () => {
    const file = input.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (e) => { preview.src = e.target?.result; };
    reader.readAsDataURL(file);
  });
};

bindImagePreview('input[name="event_image_file"]', '#event-image-preview-create');
bindImagePreview('form[action*="/edit"] input[name="event_image_file"]', '#event-image-preview-edit');
bindImagePreview('input[name="avatar_file"]', '#avatar-preview');

const bell = document.querySelector('[data-notifications-toggle]');
const panel = document.querySelector('[data-notifications-panel]');
const backdrop = document.querySelector('[data-notifications-backdrop]');
const closeBtn = document.querySelector('[data-notifications-close]');
const openNotifications = () => { panel?.classList.add('open'); backdrop?.classList.add('open'); document.body.classList.add('notifications-open'); };
const closeNotifications = () => { panel?.classList.remove('open'); backdrop?.classList.remove('open'); document.body.classList.remove('notifications-open'); };
bell?.addEventListener('click', (event) => { event.preventDefault(); event.stopPropagation(); openNotifications(); });
closeBtn?.addEventListener('click', closeNotifications);
backdrop?.addEventListener('click', closeNotifications);
document.addEventListener('keydown', (e) => { if (e.key === 'Escape') closeNotifications(); });

function setNotificationBellCount(count) {
  const bellBtn = document.querySelector('[data-notifications-toggle]');
  let dot = bellBtn?.querySelector('.notification-dot');
  if (!bellBtn) return;
  if (count <= 0) {
    bellBtn.classList.remove('has-unread');
    dot?.remove();
    return;
  }
  bellBtn.classList.add('has-unread');
  if (!dot) {
    dot = document.createElement('span');
    dot.className = 'notification-dot';
    bellBtn.appendChild(dot);
  }
  dot.textContent = count > 9 ? '9+' : String(count);
}

function updateNotificationBellCount(delta) {
  const dot = document.querySelector('[data-notifications-toggle] .notification-dot');
  if (!dot) return;
  const raw = dot.textContent.trim();
  const current = raw === '9+' ? 10 : Number(raw || 0);
  setNotificationBellCount(Math.max(0, current + delta));
}

document.querySelectorAll('[data-notification-read-form]').forEach((form) => {
  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    const response = await fetch(form.action, {
      method: 'POST',
      body: new FormData(form),
      credentials: 'same-origin',
      headers: { 'X-Requested-With': 'fetch' },
    });
    if (!response.ok) return;
    const item = form.closest('.notification-item');
    if (item) item.classList.remove('unread');
    form.remove();
    updateNotificationBellCount(-1);
  });
});

document.querySelectorAll('[data-notifications-read-all-form]').forEach((form) => {
  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    const response = await fetch(form.action, {
      method: 'POST',
      body: new FormData(form),
      credentials: 'same-origin',
      headers: { 'X-Requested-With': 'fetch' },
    });
    if (!response.ok) return;
    document.querySelectorAll('.notification-item.unread').forEach((item) => {
      item.classList.remove('unread');
      item.querySelector('form[data-notification-read-form]')?.remove();
    });
    setNotificationBellCount(0);
  });
});

function switchVolunteerSection(section, button) {
  document.querySelectorAll('.volunteer-section').forEach((el) => { el.style.display = 'none'; });
  document.querySelectorAll('.volunteer-tab').forEach((el) => { el.classList.remove('active'); });
  const target = document.getElementById(`volunteer-section-${section}`);
  if (target) target.style.display = 'block';
  if (button) button.classList.add('active');
}

document.querySelectorAll('[data-volunteer-tab]').forEach((button) => {
  button.addEventListener('click', () => switchVolunteerSection(button.dataset.volunteerTab, button));
});

function createEventRoleRow() {
  const row = document.createElement('div');
  row.className = 'event-role-row';
  row.innerHTML = `
    <input name="role_titles" placeholder="Название роли">
    <input name="role_capacities" type="number" min="1" placeholder="Кол-во людей">
    <textarea name="role_descriptions" placeholder="Кратко опишите, что будет делать волонтёр"></textarea>
    <button type="button" class="btn btn-secondary" data-remove-event-role>Удалить</button>
  `;
  return row;
}

const rolesList = document.querySelector('[data-event-roles-list]');
const addRoleBtn = document.querySelector('[data-add-event-role]');
const noRolesToggle = document.querySelector('[data-no-roles-toggle]');
const rolesBuilder = document.querySelector('[data-event-roles-builder]');
function syncNoRolesState() { if (!noRolesToggle || !rolesBuilder) return; rolesBuilder.style.display = noRolesToggle.checked ? 'none' : ''; }
addRoleBtn?.addEventListener('click', () => rolesList?.appendChild(createEventRoleRow()));
noRolesToggle?.addEventListener('change', syncNoRolesState);
syncNoRolesState();
document.addEventListener('click', (event) => {
  const removeBtn = event.target.closest('[data-remove-event-role]');
  if (!removeBtn) return;
  const row = removeBtn.closest('.event-role-row');
  if (!row) return;
  if (rolesList?.querySelectorAll('.event-role-row').length === 1) {
    row.querySelectorAll('input, textarea').forEach((el) => { el.value = ''; });
    return;
  }
  row.remove();
});

// Reusable image cropper for avatar (circle preview) and event covers (16:9 frame).
(() => {
  const cropInputs = document.querySelectorAll('input[type="file"][data-crop-kind]');
  if (!cropInputs.length) return;

  const modal = document.createElement('div');
  modal.className = 'cropper-backdrop';
  modal.innerHTML = `
    <div class="cropper-window">
      <div class="cropper-head"><h3>Обрезка изображения</h3><button type="button" class="btn btn-secondary" data-crop-close>Закрыть</button></div>
      <div class="cropper-body">
        <div class="cropper-stage" data-crop-stage><canvas class="cropper-canvas" data-crop-canvas></canvas><div class="cropper-frame" data-crop-frame></div></div>
        <aside class="cropper-side">
          <div class="cropper-preview" data-crop-preview-wrap><p class="text-muted">Предпросмотр</p><canvas data-crop-preview></canvas></div>
          <label>Масштаб</label><input data-crop-zoom type="range" min="0.5" max="3" step="0.01" value="1">
          <p class="text-muted">Можно двигать картинку мышкой/пальцем и менять масштаб, как в редакторе фото.</p>
        </aside>
      </div>
      <div class="cropper-actions"><button type="button" class="btn btn-secondary" data-crop-cancel>Отменить</button><button type="button" class="btn btn-primary" data-crop-apply>Применить</button></div>
    </div>`;
  document.body.appendChild(modal);

  const stage = modal.querySelector('[data-crop-stage]');
  const canvas = modal.querySelector('[data-crop-canvas]');
  const ctx = canvas.getContext('2d');
  const frame = modal.querySelector('[data-crop-frame]');
  const previewWrap = modal.querySelector('[data-crop-preview-wrap]');
  const previewCanvas = modal.querySelector('[data-crop-preview]');
  const previewCtx = previewCanvas.getContext('2d');
  const zoomInput = modal.querySelector('[data-crop-zoom]');
  let state = null;

  function frameRect() {
    const s = stage.getBoundingClientRect();
    const f = frame.getBoundingClientRect();
    return { x: f.left - s.left, y: f.top - s.top, w: f.width, h: f.height };
  }

  function draw() {
    if (!state) return;
    const rect = stage.getBoundingClientRect();
    canvas.width = Math.max(1, Math.round(rect.width));
    canvas.height = Math.max(1, Math.round(rect.height));
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = '#09090b';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    const cover = Math.max(canvas.width / state.img.width, canvas.height / state.img.height);
    const scale = cover * Number(zoomInput.value || 1);
    const w = state.img.width * scale;
    const h = state.img.height * scale;
    const x = (canvas.width - w) / 2 + (state.offsetX || 0);
    const y = (canvas.height - h) / 2 + (state.offsetY || 0);
    ctx.imageSmoothingQuality = 'high';
    ctx.drawImage(state.img, x, y, w, h);
    drawPreview();
  }

  function drawPreview() {
    if (!state) return;
    const fr = frameRect();
    const outW = state.kind === 'event' ? 320 : 220;
    const outH = state.kind === 'event' ? 180 : 220;
    previewCanvas.width = outW;
    previewCanvas.height = outH;
    previewCtx.clearRect(0, 0, outW, outH);
    previewCtx.drawImage(canvas, fr.x, fr.y, fr.w, fr.h, 0, 0, outW, outH);
  }

  function openCropper(input, file) {
    const img = new Image();
    img.onload = () => {
      const kind = input.dataset.cropKind || 'event';
      state = { input, img, kind, offsetX: 0, offsetY: 0 };
      frame.className = `cropper-frame ${kind}`;
      previewWrap.classList.toggle('avatar', kind === 'avatar');
      zoomInput.value = '1';
      modal.classList.add('open');
      requestAnimationFrame(draw);
    };
    img.src = URL.createObjectURL(file);
  }

  function closeCropper() { modal.classList.remove('open'); state = null; }

  function applyCrop() {
    if (!state) return;
    const fr = frameRect();
    const out = document.createElement('canvas');
    out.width = state.kind === 'event' ? 1280 : 512;
    out.height = state.kind === 'event' ? 720 : 512;
    out.getContext('2d').drawImage(canvas, fr.x, fr.y, fr.w, fr.h, 0, 0, out.width, out.height);
    const dataURL = out.toDataURL('image/png');
    const hidden = state.input.dataset.cropHidden ? document.querySelector(state.input.dataset.cropHidden) : null;
    const preview = state.input.dataset.cropPreview ? document.querySelector(state.input.dataset.cropPreview) : null;
    if (hidden) hidden.value = dataURL;
    if (preview) preview.src = dataURL;
    closeCropper();
  }

  cropInputs.forEach((input) => {
    input.addEventListener('change', () => {
      const file = input.files?.[0];
      if (file) openCropper(input, file);
    });
  });

  zoomInput.addEventListener('input', draw);
  modal.querySelector('[data-crop-apply]').addEventListener('click', applyCrop);
  modal.querySelector('[data-crop-close]').addEventListener('click', closeCropper);
  modal.querySelector('[data-crop-cancel]').addEventListener('click', closeCropper);
  modal.addEventListener('click', (event) => { if (event.target === modal) closeCropper(); });

  let drag = null;
  stage.addEventListener('pointerdown', (event) => {
    if (!state) return;
    drag = { x: event.clientX, y: event.clientY, ox: state.offsetX || 0, oy: state.offsetY || 0 };
    stage.classList.add('dragging');
    stage.setPointerCapture(event.pointerId);
  });
  stage.addEventListener('pointermove', (event) => {
    if (!drag) return;
    state.offsetX = drag.ox + event.clientX - drag.x;
    state.offsetY = drag.oy + event.clientY - drag.y;
    draw();
  });
  stage.addEventListener('pointerup', () => { drag = null; stage.classList.remove('dragging'); });
  stage.addEventListener('pointercancel', () => { drag = null; stage.classList.remove('dragging'); });
  window.addEventListener('resize', () => { if (state) draw(); });
})();
