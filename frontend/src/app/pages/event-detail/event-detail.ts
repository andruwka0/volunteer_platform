import { Component, OnInit, computed, inject, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { EventApi } from '../../entities/event/event.api';
import { EventModel, EVENT_STATUS_LABELS, EVENT_STATUS_TONE } from '../../entities/event/event.model';
import { Participant } from '../../entities/participant/participant.model';
import { ApiError } from '../../shared/api/api-error';
import { AuthStore } from '../../shared/auth/auth.store';
import { formatDateTime } from '../../shared/lib/date.util';
import { Badge, Button, Card, NotificationService, Spinner } from '../../shared/ui';

@Component({
  selector: 'cu-event-detail-page',
  imports: [RouterLink, Card, Badge, Button, Spinner],
  template: `
    <div class="event-detail">
      <!-- Back link -->
      <a routerLink="/events" class="event-detail__back">← К мероприятиям</a>

      @if (loadingEvent()) {
        <div class="event-detail__center">
          <cu-spinner label="Загрузка мероприятия…" />
        </div>
      } @else if (errorEvent()) {
        <div class="cu-alert cu-alert--error">{{ errorEvent() }}</div>
      } @else if (event()) {
        <!-- Title + status -->
        <div class="event-detail__header">
          <h1 class="event-detail__title">{{ event()!.title }}</h1>
          <cu-badge [tone]="EVENT_STATUS_TONE[event()!.status]">
            {{ EVENT_STATUS_LABELS[event()!.status] }}
          </cu-badge>
        </div>

        <!-- Main info card -->
        <cu-card class="event-detail__card">
          <p class="event-detail__description">{{ event()!.description }}</p>

          <ul class="event-detail__meta">
            <li class="event-detail__meta-item">
              <span aria-hidden="true">📍</span>
              <span>{{ event()!.location }}</span>
            </li>
            <li class="event-detail__meta-item">
              <span aria-hidden="true">🗓</span>
              <span
                >{{ formatDateTime(event()!.start_date) }} –
                {{ formatDateTime(event()!.end_date) }}</span
              >
            </li>
            @if (event()!.registration_deadline) {
              <li class="event-detail__meta-item">
                <span class="event-detail__meta-label">Регистрация до:</span>
                <span>{{ formatDateTime(event()!.registration_deadline) }}</span>
              </li>
            }
            <li class="event-detail__meta-item">
              <span aria-hidden="true">★</span>
              <span>{{ event()!.skill_points }} SP</span>
            </li>
            <li class="event-detail__meta-item">
              <span class="event-detail__meta-label">Участники:</span>
              <span>
                {{ event()!.participants_count }}
                @if (event()!.max_participants) {
                  / {{ event()!.max_participants }}
                }
              </span>
            </li>
            <li class="event-detail__meta-item">
              <span class="event-detail__meta-label">Резерв:</span>
              <span>{{ event()!.reserve_count }} / {{ event()!.reserve_participants }}</span>
            </li>
          </ul>

          <!-- Actions -->
          <div class="event-detail__actions">
            @if (
              me() && !isCreator() && event()!.status === 'EVENT-RECRUITING' && !isParticipant()
            ) {
              <cu-button [disabled]="acting()" (click)="register()">Записаться</cu-button>
            }
            @if (isParticipant()) {
              <cu-button [variant]="'outline'" [disabled]="acting()" (click)="cancel()">
                Отменить запись
              </cu-button>
            }
            @if (isAdmin()) {
              <div class="event-detail__admin-action">
                <cu-button
                  [variant]="'purple'"
                  [disabled]="acting() || event()!.status !== 'EVENT-FINISHED'"
                  (click)="approve()"
                >
                  Закрыть и начислить баллы
                </cu-button>
                @if (event()!.status !== 'EVENT-FINISHED') {
                  <span class="event-detail__hint">Доступно для завершённых мероприятий</span>
                }
              </div>
            }
          </div>
        </cu-card>

        <!-- Participants card -->
        <cu-card class="event-detail__card">
          <h2 class="event-detail__section-title">Участники</h2>

          @if (loadingParticipants()) {
            <cu-spinner label="Загрузка участников…" />
          } @else if (errorParticipants()) {
            <div class="cu-alert cu-alert--error">{{ errorParticipants() }}</div>
          } @else if (participants().length === 0) {
            <p class="event-detail__empty">Участников пока нет</p>
          } @else {
            <ul class="event-detail__participants">
              @for (p of participants(); track p.user_id) {
                <li class="event-detail__participant">
                  <div class="event-detail__participant-info">
                    <span class="event-detail__participant-name">
                      {{ participantFullName(p) }}
                    </span>
                    @if (p.telegram) {
                      <span class="event-detail__participant-tg">&#64;{{ p.telegram }}</span>
                    }
                  </div>
                  <div class="event-detail__participant-right">
                    @if (p.attendance_confirmed) {
                      <cu-badge [tone]="'success'">Подтверждён</cu-badge>
                    } @else {
                      <cu-badge [tone]="'neutral'">Не отмечен</cu-badge>
                    }
                    @if (isCreator() && !p.attendance_confirmed) {
                      <cu-button [size]="'sm'" [disabled]="acting()" (click)="confirm(p.user_id)">
                        Подтвердить посещение
                      </cu-button>
                    }
                  </div>
                </li>
              }
            </ul>
          }
        </cu-card>
      }
    </div>
  `,
  styles: `
    .event-detail {
      max-width: 860px;
      margin: 0 auto;
      padding: 24px 16px;
    }

    .event-detail__back {
      display: inline-block;
      color: var(--cu-accent);
      text-decoration: none;
      font-size: 14px;
      font-weight: 500;
      margin-bottom: 20px;
    }

    .event-detail__back:hover {
      text-decoration: underline;
    }

    .event-detail__center {
      display: flex;
      justify-content: center;
      padding: 48px 0;
    }

    .event-detail__header {
      display: flex;
      align-items: flex-start;
      gap: 12px;
      flex-wrap: wrap;
      margin-bottom: 20px;
    }

    .event-detail__title {
      font-size: 28px;
      font-weight: 700;
      color: var(--cu-text);
      margin: 0;
      flex: 1;
      min-width: 0;
    }

    .event-detail__card {
      margin-bottom: 20px;
    }

    .event-detail__description {
      color: var(--cu-text);
      line-height: 1.6;
      margin: 0 0 20px;
    }

    .event-detail__meta {
      list-style: none;
      padding: 0;
      margin: 0 0 24px;
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .event-detail__meta-item {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 14px;
      color: var(--cu-text-2);
    }

    .event-detail__meta-label {
      font-weight: 600;
      color: var(--cu-text);
    }

    .event-detail__actions {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 12px;
    }

    .event-detail__admin-action {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .event-detail__hint {
      font-size: 12px;
      color: var(--cu-text-2);
    }

    .event-detail__section-title {
      font-size: 18px;
      font-weight: 700;
      color: var(--cu-text);
      margin: 0 0 16px;
    }

    .event-detail__empty {
      color: var(--cu-text-2);
      font-size: 14px;
      text-align: center;
      padding: 24px 0;
    }

    .event-detail__participants {
      list-style: none;
      padding: 0;
      margin: 0;
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .event-detail__participant {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      padding: 12px 0;
      border-bottom: 1px solid var(--cu-border);
      flex-wrap: wrap;
    }

    .event-detail__participant:last-child {
      border-bottom: none;
    }

    .event-detail__participant-info {
      display: flex;
      flex-direction: column;
      gap: 2px;
      min-width: 0;
    }

    .event-detail__participant-name {
      font-size: 14px;
      font-weight: 600;
      color: var(--cu-text);
    }

    .event-detail__participant-tg {
      font-size: 13px;
      color: var(--cu-text-2);
    }

    .event-detail__participant-right {
      display: flex;
      align-items: center;
      gap: 8px;
      flex-wrap: wrap;
    }
  `,
})
export class EventDetailPage implements OnInit {
  readonly id = input.required<string>();

  private readonly eventApi = inject(EventApi);
  private readonly store = inject(AuthStore);
  private readonly notifications = inject(NotificationService);

  protected readonly loadingEvent = signal(true);
  protected readonly errorEvent = signal<string | null>(null);
  protected readonly event = signal<EventModel | null>(null);

  protected readonly loadingParticipants = signal(true);
  protected readonly errorParticipants = signal<string | null>(null);
  protected readonly participants = signal<Participant[]>([]);

  protected readonly acting = signal(false);

  protected readonly me = this.store.user;
  protected readonly isAdmin = this.store.isAdmin;

  protected readonly isParticipant = computed(() => {
    const me = this.me();
    if (!me) return false;
    return this.participants().some((p) => p.user_id === me.id);
  });

  protected readonly isCreator = computed(() => {
    const me = this.me();
    const ev = this.event();
    if (!me || !ev) return false;
    return ev.created_by_id === me.id;
  });

  protected readonly EVENT_STATUS_LABELS = EVENT_STATUS_LABELS;
  protected readonly EVENT_STATUS_TONE = EVENT_STATUS_TONE;
  protected readonly formatDateTime = formatDateTime;

  private get numId(): number {
    return Number(this.id());
  }

  private loadEvent(): void {
    this.loadingEvent.set(true);
    this.errorEvent.set(null);
    this.eventApi.byId(this.numId).subscribe({
      next: (ev) => {
        this.event.set(ev);
        this.loadingEvent.set(false);
      },
      error: (e: ApiError) => {
        this.errorEvent.set(e.message);
        this.loadingEvent.set(false);
      },
    });
  }

  private loadParticipants(): void {
    this.loadingParticipants.set(true);
    this.errorParticipants.set(null);
    this.eventApi.participants(this.numId).subscribe({
      next: (res) => {
        this.participants.set(res.participants);
        this.loadingParticipants.set(false);
      },
      error: (e: ApiError) => {
        this.errorParticipants.set(e.message);
        this.loadingParticipants.set(false);
      },
    });
  }

  private load(): void {
    this.loadEvent();
    this.loadParticipants();
  }

  ngOnInit(): void {
    this.load();
  }

  protected participantFullName(p: Participant): string {
    return [p.last_name, p.first_name, p.middle_name].filter(Boolean).join(' ').trim();
  }

  protected register(): void {
    this.acting.set(true);
    this.eventApi.register(this.numId).subscribe({
      next: (res) => {
        this.notifications.success(res.message);
        this.acting.set(false);
        this.load();
      },
      error: (e: ApiError) => {
        this.notifications.error(e.message);
        this.acting.set(false);
      },
    });
  }

  protected cancel(): void {
    this.acting.set(true);
    this.eventApi.cancelRegistration(this.numId).subscribe({
      next: (res) => {
        this.notifications.success(res.message);
        this.acting.set(false);
        this.load();
      },
      error: (e: ApiError) => {
        this.notifications.error(e.message);
        this.acting.set(false);
      },
    });
  }

  protected approve(): void {
    this.acting.set(true);
    this.eventApi.approve(this.numId).subscribe({
      next: (res) => {
        this.notifications.success(res.message);
        this.acting.set(false);
        this.load();
      },
      error: (e: ApiError) => {
        this.notifications.error(e.message);
        this.acting.set(false);
      },
    });
  }

  protected confirm(userId: number): void {
    this.acting.set(true);
    this.eventApi.confirmAttendance(this.numId, userId).subscribe({
      next: (res) => {
        this.notifications.success(res.message);
        this.acting.set(false);
        this.load();
      },
      error: (e: ApiError) => {
        this.notifications.error(e.message);
        this.acting.set(false);
      },
    });
  }
}
