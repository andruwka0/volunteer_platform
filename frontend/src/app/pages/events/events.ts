import { Component, computed, inject, signal } from '@angular/core';
import { EventApi } from '../../entities/event/event.api';
import { EventModel, EventStatus } from '../../entities/event/event.model';
import { ApiError } from '../../shared/api/api-error';
import { EmptyState, PageHeader, PillTab, PillTabs, Spinner } from '../../shared/ui';
import { EventCard } from '../../widgets/event-card/event-card';

@Component({
  selector: 'cu-events-page',
  imports: [PageHeader, PillTabs, EventCard, Spinner, EmptyState],
  template: `
    <div class="events-page">
      <cu-page-header title="Мероприятия" subtitle="Выберите мероприятие и запишитесь" />

      <div class="events-page__filters">
        <cu-pill-tabs [tabs]="tabs" [(value)]="filter" />
      </div>

      @if (loading()) {
        <div class="events-page__center">
          <cu-spinner label="Загрузка мероприятий…" />
        </div>
      } @else if (error()) {
        <div class="cu-alert cu-alert--error">{{ error() }}</div>
      } @else {
        @if (filtered().length === 0) {
          <cu-empty-state
            icon="📅"
            title="Мероприятий не найдено"
            text="По выбранному фильтру нет доступных мероприятий"
          />
        } @else {
          <div class="events-page__grid">
            @for (e of filtered(); track e.id) {
              <cu-event-card [event]="e" />
            }
          </div>
        }
      }
    </div>
  `,
  styles: `
    .events-page {
      max-width: 1200px;
      margin: 0 auto;
      padding: 24px 16px;
    }

    .events-page__filters {
      margin: 16px 0 24px;
    }

    .events-page__center {
      display: flex;
      justify-content: center;
      padding: 48px 0;
    }

    .events-page__grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
      gap: 16px;
    }
  `,
})
export class EventsPage {
  private readonly eventApi = inject(EventApi);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly events = signal<EventModel[]>([]);

  protected readonly filter = signal<string>('');

  protected readonly tabs: PillTab[] = [
    { value: '', label: 'Все' },
    { value: 'EVENT-RECRUITING', label: 'Набор открыт' },
    { value: 'EVENT-ACTIVE', label: 'Идёт' },
    { value: 'EVENT-FINISHED', label: 'Завершён' },
    { value: 'EVENT-CLOSED', label: 'Закрыт' },
    { value: 'EVENT-CANCELLED', label: 'Отменён' },
  ];

  protected readonly filtered = computed<EventModel[]>(() => {
    const f = this.filter();
    const all = this.events();
    if (!f) return all;
    return all.filter((e) => e.status === (f as EventStatus));
  });

  private load(): void {
    this.loading.set(true);
    this.error.set(null);
    this.eventApi.list().subscribe({
      next: (res) => {
        this.events.set(res.events);
        this.loading.set(false);
      },
      error: (e: ApiError) => {
        this.error.set(e.message);
        this.loading.set(false);
      },
    });
  }

  constructor() {
    this.load();
  }
}
