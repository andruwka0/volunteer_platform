import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiError } from '../../shared/api/api-error';
import { ImagesApi } from '../../shared/api/images.api';
import { AuthStore } from '../../shared/auth/auth.store';
import { toRFC3339, formatDuration } from '../../shared/lib/date.util';
import {
  Button,
  Card,
  EmptyState,
  NotificationService,
  PageHeader,
  PillTab,
  PillTabs,
  Spinner,
} from '../../shared/ui';
import { BlacklistApi } from '../../entities/blacklist/blacklist.api';
import { EventApi } from '../../entities/event/event.api';
import { CreateEventPayload } from '../../entities/event/event.model';
import { EventTemplateApi } from '../../entities/event-template/event-template.api';
import {
  CreateEventTemplatePayload,
  EventTemplate,
} from '../../entities/event-template/event-template.model';

@Component({
  selector: 'cu-organizer-page',
  imports: [
    ReactiveFormsModule,
    PageHeader,
    PillTabs,
    Button,
    Card,
    Spinner,
    EmptyState,
  ],
  template: `
    <div class="org-page">
      <cu-page-header title="Организатору" />

      <div class="org-page__tabs">
        <cu-pill-tabs [tabs]="tabs" [(value)]="tab" />
      </div>

      <!-- ─── A) Создать мероприятие ─── -->
      @if (tab() === 'create-event') {
        <section class="org-section">
          <h2 class="org-section__title">Создать мероприятие</h2>
          <cu-card>
            <form [formGroup]="eventForm" (ngSubmit)="submitEvent()" class="cu-form">

              <div class="cu-field">
                <label class="cu-label" for="ev-title">Название *</label>
                <input id="ev-title" class="cu-input" type="text" formControlName="title"
                  placeholder="Название мероприятия" />
                @if (eventForm.get('title')?.invalid && eventForm.get('title')?.touched) {
                  <span class="cu-field__error">Обязательное поле</span>
                }
              </div>

              <div class="cu-field">
                <label class="cu-label" for="ev-description">Описание</label>
                <textarea id="ev-description" class="cu-textarea" formControlName="description"
                  rows="3" placeholder="Описание мероприятия"></textarea>
              </div>

              <div class="cu-field">
                <label class="cu-label" for="ev-location">Место проведения</label>
                <input id="ev-location" class="cu-input" type="text" formControlName="location"
                  placeholder="Адрес или название места" />
              </div>

              <div class="cu-field">
                <label class="cu-label" for="ev-image">Обложка</label>
                @if (imagesLoading()) {
                  <cu-spinner [size]="20" />
                } @else {
                  <select id="ev-image" class="cu-select" formControlName="image">
                    <option value="">— без обложки —</option>
                    @for (img of images(); track img) {
                      <option [value]="img">{{ img }}</option>
                    }
                  </select>
                }
              </div>

              <div class="cu-field">
                <label class="cu-label" for="ev-start">Начало *</label>
                <input id="ev-start" class="cu-input" type="datetime-local" formControlName="start_date" />
                @if (eventForm.get('start_date')?.invalid && eventForm.get('start_date')?.touched) {
                  <span class="cu-field__error">Обязательное поле</span>
                }
              </div>

              <div class="cu-field">
                <label class="cu-label" for="ev-end">Конец *</label>
                <input id="ev-end" class="cu-input" type="datetime-local" formControlName="end_date" />
                @if (eventForm.get('end_date')?.invalid && eventForm.get('end_date')?.touched) {
                  <span class="cu-field__error">Обязательное поле</span>
                }
              </div>

              <div class="cu-field">
                <label class="cu-label" for="ev-deadline">Дедлайн регистрации</label>
                <input id="ev-deadline" class="cu-input" type="datetime-local"
                  formControlName="registration_deadline" />
              </div>

              <div class="cu-fields-row">
                <div class="cu-field">
                  <label class="cu-label" for="ev-max">Макс. участников</label>
                  <input id="ev-max" class="cu-input" type="number" min="1"
                    formControlName="max_participants" placeholder="Без ограничений" />
                </div>
                <div class="cu-field">
                  <label class="cu-label" for="ev-reserve">Резерв</label>
                  <input id="ev-reserve" class="cu-input" type="number" min="0"
                    formControlName="reserve_participants" />
                </div>
                <div class="cu-field">
                  <label class="cu-label" for="ev-sp">Баллы</label>
                  <input id="ev-sp" class="cu-input" type="number" min="0"
                    formControlName="skill_points" />
                </div>
              </div>

              <div class="cu-field">
                <label class="cu-label" for="ev-template">Шаблон</label>
                @if (templatesLoading()) {
                  <cu-spinner [size]="20" />
                } @else {
                  <select id="ev-template" class="cu-select" formControlName="template_id">
                    <option value="">— без шаблона —</option>
                    @for (t of templates(); track t.id) {
                      <option [value]="t.id">{{ t.title }}</option>
                    }
                  </select>
                }
              </div>

              <div class="cu-form__actions">
                <cu-button type="submit" [variant]="'primary'"
                  [loading]="eventSubmitting()"
                  [disabled]="eventForm.invalid || eventSubmitting()">
                  Создать мероприятие
                </cu-button>
              </div>
            </form>
          </cu-card>
        </section>
      }

      <!-- ─── B) Шаблоны ─── -->
      @if (tab() === 'templates') {
        <section class="org-section">
          <h2 class="org-section__title">Шаблоны мероприятий</h2>

          @if (templatesLoading()) {
            <div class="org-center"><cu-spinner label="Загрузка шаблонов…" /></div>
          } @else if (templatesError()) {
            <div class="cu-alert cu-alert--error">{{ templatesError() }}</div>
          } @else if (templates().length === 0) {
            <cu-empty-state icon="📋" title="Нет шаблонов" text="Создайте первый шаблон ниже" />
          } @else {
            <div class="org-grid">
              @for (t of templates(); track t.id) {
                <cu-card>
                  <p class="org-card__title">{{ t.title }}</p>
                  @if (t.location) {
                    <p class="org-card__meta">📍 {{ t.location }}</p>
                  }
                  <p class="org-card__meta">⏱ {{ formatDuration(t.duration_minutes) }}</p>
                  <p class="org-card__meta">★ {{ t.skill_points }} баллов</p>
                  <p class="org-card__meta">
                    Участники: до {{ t.max_participants ?? '∞' }}
                    / резерв {{ t.reserve_participants }}
                  </p>
                </cu-card>
              }
            </div>
          }

          <h3 class="org-section__subtitle">Создать шаблон</h3>
          <cu-card>
            <form [formGroup]="templateForm" (ngSubmit)="submitTemplate()" class="cu-form">

              <div class="cu-field">
                <label class="cu-label" for="tpl-title">Название *</label>
                <input id="tpl-title" class="cu-input" type="text" formControlName="title"
                  placeholder="Название шаблона" />
                @if (templateForm.get('title')?.invalid && templateForm.get('title')?.touched) {
                  <span class="cu-field__error">Обязательное поле</span>
                }
              </div>

              <div class="cu-field">
                <label class="cu-label" for="tpl-desc">Описание</label>
                <textarea id="tpl-desc" class="cu-textarea" formControlName="description"
                  rows="3"></textarea>
              </div>

              <div class="cu-field">
                <label class="cu-label" for="tpl-location">Место</label>
                <input id="tpl-location" class="cu-input" type="text" formControlName="location" />
              </div>

              <div class="cu-field">
                <label class="cu-label" for="tpl-image">URL обложки</label>
                <input id="tpl-image" class="cu-input" type="text" formControlName="image_url"
                  placeholder="https://…" />
              </div>

              <div class="cu-fields-row">
                <div class="cu-field">
                  <label class="cu-label" for="tpl-dur">Длительность (мин) *</label>
                  <input id="tpl-dur" class="cu-input" type="number" min="1"
                    formControlName="duration_minutes" />
                  @if (templateForm.get('duration_minutes')?.invalid && templateForm.get('duration_minutes')?.touched) {
                    <span class="cu-field__error">Укажите длительность</span>
                  }
                </div>
                <div class="cu-field">
                  <label class="cu-label" for="tpl-max">Макс. участников</label>
                  <input id="tpl-max" class="cu-input" type="number" min="1"
                    formControlName="max_participants" />
                </div>
                <div class="cu-field">
                  <label class="cu-label" for="tpl-reserve">Резерв</label>
                  <input id="tpl-reserve" class="cu-input" type="number" min="0"
                    formControlName="reserve_participants" />
                </div>
                <div class="cu-field">
                  <label class="cu-label" for="tpl-sp">Баллы</label>
                  <input id="tpl-sp" class="cu-input" type="number" min="0"
                    formControlName="skill_points" />
                </div>
              </div>

              <div class="cu-form__actions">
                <cu-button type="submit" [variant]="'primary'"
                  [loading]="templateSubmitting()"
                  [disabled]="templateForm.invalid || templateSubmitting()">
                  Создать шаблон
                </cu-button>
              </div>
            </form>
          </cu-card>
        </section>
      }

      <!-- ─── C) Чёрный список ─── -->
      @if (tab() === 'blacklist') {
        <section class="org-section">
          <h2 class="org-section__title">Чёрный список</h2>
          <cu-card>
            <form [formGroup]="blacklistForm" class="cu-form">
              <div class="cu-field">
                <label class="cu-label" for="bl-uid">ID пользователя</label>
                <input id="bl-uid" class="cu-input" type="number" min="1"
                  formControlName="user_id" placeholder="Введите ID" />
                @if (blacklistForm.get('user_id')?.invalid && blacklistForm.get('user_id')?.touched) {
                  <span class="cu-field__error">Введите корректный ID</span>
                }
              </div>
              <div class="cu-form__actions cu-form__actions--gap">
                <cu-button [variant]="'danger'"
                  [loading]="banLoading()"
                  [disabled]="blacklistForm.invalid || banLoading() || unbanLoading()"
                  (click)="ban()">
                  Забанить
                </cu-button>
                <cu-button [variant]="'outline'"
                  [loading]="unbanLoading()"
                  [disabled]="blacklistForm.invalid || banLoading() || unbanLoading()"
                  (click)="unban()">
                  Разбанить
                </cu-button>
              </div>
            </form>
          </cu-card>
        </section>
      }
    </div>
  `,
  styles: `
    .org-page {
      max-width: 960px;
      margin: 0 auto;
      padding: 24px 16px;
    }
    .org-page__tabs {
      margin: 16px 0 24px;
    }
    .org-section {
      display: flex;
      flex-direction: column;
      gap: 20px;
    }
    .org-section__title {
      font-size: 20px;
      font-weight: 700;
      color: var(--cu-text);
      margin: 0;
    }
    .org-section__subtitle {
      font-size: 17px;
      font-weight: 600;
      color: var(--cu-text);
      margin: 8px 0 0;
    }
    .org-center {
      display: flex;
      justify-content: center;
      padding: 32px 0;
    }
    .org-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
      gap: 16px;
    }
    .org-card__title {
      font-weight: 700;
      font-size: 15px;
      margin: 0 0 8px;
      color: var(--cu-text);
    }
    .org-card__meta {
      font-size: 13px;
      color: var(--cu-text-secondary);
      margin: 2px 0;
    }
    .cu-form {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
    .cu-fields-row {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
    }
    .cu-fields-row .cu-field {
      flex: 1;
      min-width: 120px;
    }
    .cu-form__actions {
      display: flex;
      justify-content: flex-end;
      padding-top: 4px;
    }
    .cu-form__actions--gap {
      gap: 12px;
    }
  `,
})
export class OrganizerPage {
  private readonly eventApi = inject(EventApi);
  private readonly eventTemplateApi = inject(EventTemplateApi);
  private readonly blacklistApi = inject(BlacklistApi);
  private readonly imagesApi = inject(ImagesApi);
  private readonly notif = inject(NotificationService);
  private readonly fb = inject(FormBuilder);

  protected readonly store = inject(AuthStore);

  protected readonly tab = signal('create-event');
  protected readonly tabs: PillTab[] = [
    { value: 'create-event', label: 'Создать мероприятие' },
    { value: 'templates', label: 'Шаблоны' },
    { value: 'blacklist', label: 'Чёрный список' },
  ];

  // Images
  protected readonly images = signal<string[]>([]);
  protected readonly imagesLoading = signal(true);

  // Templates
  protected readonly templates = signal<EventTemplate[]>([]);
  protected readonly templatesLoading = signal(true);
  protected readonly templatesError = signal<string | null>(null);

  // Form submission states
  protected readonly eventSubmitting = signal(false);
  protected readonly templateSubmitting = signal(false);
  protected readonly banLoading = signal(false);
  protected readonly unbanLoading = signal(false);

  protected readonly formatDuration = formatDuration;

  // ── Reactive Forms ──────────────────────────────────────

  protected readonly eventForm = this.fb.group({
    title: ['', Validators.required],
    description: [''],
    location: [''],
    image: [''],
    start_date: ['', Validators.required],
    end_date: ['', Validators.required],
    registration_deadline: [''],
    max_participants: [null as number | null],
    reserve_participants: [0],
    skill_points: [0],
    template_id: [''],
  });

  protected readonly templateForm = this.fb.group({
    title: ['', Validators.required],
    description: [''],
    location: [''],
    image_url: [''],
    duration_minutes: [null as number | null, [Validators.required, Validators.min(1)]],
    max_participants: [null as number | null],
    reserve_participants: [0],
    skill_points: [0],
  });

  protected readonly blacklistForm = this.fb.group({
    user_id: [null as number | null, [Validators.required, Validators.min(1)]],
  });

  // ── Lifecycle ───────────────────────────────────────────

  constructor() {
    this.loadImages();
    this.loadTemplates();
  }

  private loadImages(): void {
    this.imagesLoading.set(true);
    this.imagesApi.list().subscribe({
      next: (res) => {
        this.images.set(res.images);
        this.imagesLoading.set(false);
      },
      error: () => {
        this.images.set([]);
        this.imagesLoading.set(false);
      },
    });
  }

  private loadTemplates(): void {
    this.templatesLoading.set(true);
    this.templatesError.set(null);
    this.eventTemplateApi.list().subscribe({
      next: (res) => {
        this.templates.set(res.templates);
        this.templatesLoading.set(false);
      },
      error: (e: ApiError) => {
        this.templatesError.set(e.message);
        this.templatesLoading.set(false);
      },
    });
  }

  // ── Submit: Create Event ────────────────────────────────

  protected submitEvent(): void {
    if (this.eventForm.invalid) {
      this.eventForm.markAllAsTouched();
      return;
    }

    const v = this.eventForm.getRawValue();

    const startRFC = toRFC3339(v.start_date);
    const endRFC = toRFC3339(v.end_date);
    if (!startRFC || !endRFC) {
      this.notif.error('Укажите корректные даты начала и конца');
      return;
    }

    const templateIdNum = v.template_id ? Number(v.template_id) : null;

    const payload: CreateEventPayload = {
      title: v.title ?? '',
      description: v.description ?? '',
      location: v.location ?? '',
      image: v.image ?? '',
      start_date: startRFC,
      end_date: endRFC,
      registration_deadline: toRFC3339(v.registration_deadline) ?? null,
      max_participants: v.max_participants ?? null,
      reserve_participants: v.reserve_participants ?? 0,
      skill_points: v.skill_points ?? 0,
      template_id: templateIdNum,
    };

    this.eventSubmitting.set(true);
    this.eventApi.create(payload).subscribe({
      next: () => {
        this.notif.success('Мероприятие успешно создано');
        this.eventForm.reset({
          title: '',
          description: '',
          location: '',
          image: '',
          start_date: '',
          end_date: '',
          registration_deadline: '',
          max_participants: null,
          reserve_participants: 0,
          skill_points: 0,
          template_id: '',
        });
        this.eventSubmitting.set(false);
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
        this.eventSubmitting.set(false);
      },
    });
  }

  // ── Submit: Create Template ─────────────────────────────

  protected submitTemplate(): void {
    if (this.templateForm.invalid) {
      this.templateForm.markAllAsTouched();
      return;
    }

    const v = this.templateForm.getRawValue();

    const payload: CreateEventTemplatePayload = {
      title: v.title ?? '',
      description: v.description ?? '',
      location: v.location ?? '',
      image_url: v.image_url ?? '',
      duration_minutes: v.duration_minutes ?? 0,
      max_participants: v.max_participants ?? null,
      reserve_participants: v.reserve_participants ?? 0,
      skill_points: v.skill_points ?? 0,
    };

    this.templateSubmitting.set(true);
    this.eventTemplateApi.create(payload).subscribe({
      next: () => {
        this.notif.success('Шаблон успешно создан');
        this.templateForm.reset({
          title: '',
          description: '',
          location: '',
          image_url: '',
          duration_minutes: null,
          max_participants: null,
          reserve_participants: 0,
          skill_points: 0,
        });
        this.templateSubmitting.set(false);
        this.loadTemplates();
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
        this.templateSubmitting.set(false);
      },
    });
  }

  // ── Blacklist ───────────────────────────────────────────

  protected ban(): void {
    if (this.blacklistForm.invalid) {
      this.blacklistForm.markAllAsTouched();
      return;
    }
    const userId = this.blacklistForm.getRawValue().user_id;
    if (!userId) return;

    this.banLoading.set(true);
    this.blacklistApi.ban(userId).subscribe({
      next: () => {
        this.notif.success(`Пользователь ${userId} заблокирован`);
        this.banLoading.set(false);
        this.blacklistForm.reset({ user_id: null });
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
        this.banLoading.set(false);
      },
    });
  }

  protected unban(): void {
    if (this.blacklistForm.invalid) {
      this.blacklistForm.markAllAsTouched();
      return;
    }
    const userId = this.blacklistForm.getRawValue().user_id;
    if (!userId) return;

    this.unbanLoading.set(true);
    this.blacklistApi.unban(userId).subscribe({
      next: () => {
        this.notif.success(`Пользователь ${userId} разблокирован`);
        this.unbanLoading.set(false);
        this.blacklistForm.reset({ user_id: null });
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
        this.unbanLoading.set(false);
      },
    });
  }
}
