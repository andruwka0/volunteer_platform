import { Component, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { EventModel, EVENT_STATUS_LABELS, EVENT_STATUS_TONE } from '../../entities/event/event.model';
import { formatDateTime } from '../../shared/lib/date.util';
import { Badge, Card } from '../../shared/ui';

@Component({
  selector: 'cu-event-card',
  imports: [RouterLink, Card, Badge],
  template: `
    <a [routerLink]="['/events', event().id]" class="event-card-link">
      <cu-card [interactive]="true">
        <!-- Cover image area -->
        <div class="event-card__cover">
          @if (event().cover_image_url && !imgFailed()) {
            <img
              [src]="event().cover_image_url"
              [alt]="event().title"
              class="event-card__img"
              (error)="imgFailed.set(true)"
            />
          } @else {
            <div class="event-card__placeholder" aria-hidden="true">
              <span class="event-card__placeholder-letter">{{ event().title.charAt(0) }}</span>
            </div>
          }
        </div>

        <!-- Card body -->
        <div class="event-card__body">
          <!-- Status badge -->
          <div class="event-card__status-row">
            <cu-badge [tone]="EVENT_STATUS_TONE[event().status]">
              {{ EVENT_STATUS_LABELS[event().status] }}
            </cu-badge>
          </div>

          <!-- Title -->
          <h3 class="event-card__title">{{ event().title }}</h3>

          <!-- Meta -->
          <ul class="event-card__meta">
            <li class="event-card__meta-item">
              <span aria-hidden="true">📍</span>
              <span>{{ event().location }}</span>
            </li>
            <li class="event-card__meta-item">
              <span aria-hidden="true">🗓</span>
              <span>{{ formatDateTime(event().start_date) }}</span>
            </li>
            <li class="event-card__meta-item">
              <span aria-hidden="true">★</span>
              <span>{{ event().skill_points }} SP</span>
            </li>
            <li class="event-card__meta-item">
              <span aria-hidden="true">👥</span>
              <span>
                {{ event().participants_count }}
                @if (event().max_participants) {
                  / {{ event().max_participants }}
                }
              </span>
            </li>
          </ul>
        </div>
      </cu-card>
    </a>
  `,
  styles: `
    .event-card-link {
      display: block;
      text-decoration: none;
      color: inherit;
      height: 100%;
    }

    .event-card-link cu-card {
      display: block;
      height: 100%;
      overflow: hidden;
    }

    .event-card__cover {
      height: 140px;
      overflow: hidden;
      border-radius: var(--cu-radius-lg, 16px) var(--cu-radius-lg, 16px) 0 0;
      margin: calc(var(--cu-space-4, 16px) * -1) calc(var(--cu-space-4, 16px) * -1) 0;
    }

    .event-card__img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }

    .event-card__placeholder {
      width: 100%;
      height: 100%;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, var(--cu-accent, #3b82f6), var(--cu-purple, #7c3aed));
    }

    .event-card__placeholder-letter {
      font-size: 48px;
      font-weight: 700;
      color: #fff;
      text-transform: uppercase;
      line-height: 1;
    }

    .event-card__body {
      padding: 12px 0 0;
    }

    .event-card__status-row {
      margin-bottom: 8px;
    }

    .event-card__title {
      font-size: 15px;
      font-weight: 700;
      color: var(--cu-text);
      margin: 0 0 10px;
      line-height: 1.4;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }

    .event-card__meta {
      list-style: none;
      margin: 0;
      padding: 0;
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .event-card__meta-item {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 13px;
      color: var(--cu-text-2);
    }
  `,
})
export class EventCard {
  readonly event = input.required<EventModel>();
  protected readonly imgFailed = signal(false);

  protected readonly EVENT_STATUS_LABELS = EVENT_STATUS_LABELS;
  protected readonly EVENT_STATUS_TONE = EVENT_STATUS_TONE;
  protected readonly formatDateTime = formatDateTime;
}
