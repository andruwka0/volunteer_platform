import { Component, inject, signal } from '@angular/core';
import { RewardApi } from '../../entities/reward/reward.api';
import { UserReward, UserRewardList } from '../../entities/reward/reward.model';
import { EventStatus, EVENT_STATUS_LABELS, EVENT_STATUS_TONE } from '../../entities/event/event.model';
import { UserEventHistoryList } from '../../entities/participant/participant.model';
import { SkillPointHistory, TRANSACTION_TYPE_LABELS } from '../../entities/skill-transaction/skill-transaction.model';
import { MeApi } from '../../entities/user/me.api';
import { UserApi } from '../../entities/user/user.api';
import { fullName, initials, ROLE_LABELS } from '../../entities/user/user.model';
import { ApiError } from '../../shared/api/api-error';
import { AuthStore } from '../../shared/auth/auth.store';
import { formatDate, formatDateTime } from '../../shared/lib/date.util';
import { Avatar, Badge, Button, Card, NotificationService, PageHeader, Spinner } from '../../shared/ui';

@Component({
  selector: 'cu-profile-page',
  imports: [PageHeader, Card, Avatar, Badge, Button, Spinner],
  template: `
    <cu-page-header title="Профиль" />

    <div class="profile-layout">

      <!-- 1. Summary -->
      <cu-card>
        <h2 class="section-title">Обо мне</h2>
        @if (store.user(); as user) {
          <div class="summary-row">
            <cu-avatar [initials]="getInitials(user)" [size]="64" />
            <div class="summary-info">
              <p class="summary-name">{{ getFullName(user) }}</p>
              <p class="summary-login">&#64;{{ user.login }}</p>
              <div class="summary-meta">
                <cu-badge [tone]="'purple'">{{ ROLE_LABELS[user.role] }}</cu-badge>
                @if (user.telegram) {
                  <span class="meta-telegram">Telegram: {{ user.telegram }}</span>
                }
              </div>
            </div>
            <div class="summary-sp">
              <span class="sp-number">{{ store.skillPoints() }}</span>
              <span class="sp-label">SP</span>
            </div>
          </div>
        }
      </cu-card>

      <!-- 2. My rewards -->
      <cu-card>
        <h2 class="section-title">Мои награды</h2>
        @if (rewardsLoading()) {
          <cu-spinner label="Загружаем награды..." />
        } @else if (rewardsError()) {
          <div class="cu-alert cu-alert--error">{{ rewardsError() }}</div>
        } @else if (myRewards().rewards.length === 0) {
          <p class="empty-text">У вас пока нет наград.</p>
        } @else {
          <div class="reward-list">
            @for (x of myRewards().rewards; track x.reward_id) {
              <div class="reward-row">
                <div class="reward-info">
                  <span class="reward-name">{{ x.name }}</span>
                  <span class="reward-cost">&#9733; {{ x.cost }} SP</span>
                </div>
                <div class="reward-status">
                  @if (x.picked_up) {
                    <cu-badge [tone]="'success'">Получено</cu-badge>
                  } @else if (x.claimed && !x.picked_up) {
                    <cu-badge [tone]="'warning'">Ожидает выдачи</cu-badge>
                  } @else if (x.available && !x.claimed) {
                    <cu-button [size]="'sm'" (click)="claimReward(x)">
                      Получить за {{ x.cost }} SP
                    </cu-button>
                  }
                </div>
              </div>
            }
          </div>
        }
      </cu-card>

      <!-- 3. Event participation history -->
      <cu-card>
        <h2 class="section-title">История участия</h2>
        @if (eventsLoading()) {
          <cu-spinner label="Загружаем историю..." />
        } @else if (eventsError()) {
          <div class="cu-alert cu-alert--error">{{ eventsError() }}</div>
        } @else if (eventHistory().events.length === 0) {
          <p class="empty-text">Вы ещё не участвовали в мероприятиях.</p>
        } @else {
          <div class="table-wrap">
            <table class="data-table" aria-label="История участия">
              <thead>
                <tr>
                  <th>Мероприятие</th>
                  <th>Статус</th>
                  <th>Дата записи</th>
                  <th>Посещение</th>
                  <th>Баллы</th>
                </tr>
              </thead>
              <tbody>
                @for (ev of eventHistory().events; track ev.event_id) {
                  <tr>
                    <td>{{ ev.title }}</td>
                    <td>
                      <cu-badge [tone]="getEventStatusTone(ev.status)">
                        {{ getEventStatusLabel(ev.status) }}
                      </cu-badge>
                    </td>
                    <td>{{ formatDate(ev.joined_at) }}</td>
                    <td>
                      @if (ev.attendance_confirmed) {
                        <cu-badge [tone]="'success'">Подтверждён</cu-badge>
                      } @else {
                        <cu-badge [tone]="'neutral'">Не отмечен</cu-badge>
                      }
                    </td>
                    <td class="points-cell">&#9733; {{ ev.skill_points }} SP</td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        }
      </cu-card>

      <!-- 4. Skill point history -->
      <cu-card>
        <h2 class="section-title">История баллов</h2>
        @if (txLoading()) {
          <cu-spinner label="Загружаем историю баллов..." />
        } @else if (txError()) {
          <div class="cu-alert cu-alert--error">{{ txError() }}</div>
        } @else if (txHistory().transactions.length === 0) {
          <p class="empty-text">История баллов пуста.</p>
        } @else {
          <div class="table-wrap">
            <table class="data-table" aria-label="История баллов">
              <thead>
                <tr>
                  <th>Баллы</th>
                  <th>Тип</th>
                  <th>Причина</th>
                  <th>Дата</th>
                </tr>
              </thead>
              <tbody>
                @for (tx of txHistory().transactions; track tx.ID) {
                  <tr>
                    <td>
                      <span
                        [class.points-positive]="tx.Points > 0"
                        [class.points-negative]="tx.Points < 0"
                        class="points-value"
                      >
                        {{ tx.Points > 0 ? '+' : '' }}{{ tx.Points }}
                      </span>
                    </td>
                    <td>{{ TRANSACTION_TYPE_LABELS[tx.Type] ?? tx.Type }}</td>
                    <td>{{ tx.Reason || '—' }}</td>
                    <td>{{ formatDateTime(tx.CreatedAt) }}</td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        }
      </cu-card>

    </div>
  `,
  styles: `
    :host {
      display: block;
      padding: 24px;
    }
    .profile-layout {
      display: flex;
      flex-direction: column;
      gap: 20px;
      margin-top: 24px;
    }
    .section-title {
      font-size: 18px;
      font-weight: 700;
      color: var(--cu-ink);
      margin: 0 0 16px;
    }
    /* Summary section */
    .summary-row {
      display: flex;
      align-items: flex-start;
      gap: 20px;
    }
    .summary-info {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }
    .summary-name {
      font-size: 20px;
      font-weight: 700;
      color: var(--cu-ink);
      margin: 0;
    }
    .summary-login {
      font-size: 14px;
      color: var(--cu-text-muted);
      margin: 0;
    }
    .summary-meta {
      display: flex;
      align-items: center;
      gap: 10px;
      flex-wrap: wrap;
    }
    .meta-telegram {
      font-size: 13px;
      color: var(--cu-text-muted);
    }
    .summary-sp {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      min-width: 80px;
      padding: 12px;
      background: var(--cu-accent-soft, #e8f0fe);
      border-radius: var(--cu-radius);
    }
    .sp-number {
      font-size: 32px;
      font-weight: 800;
      color: var(--cu-accent);
      line-height: 1;
    }
    .sp-label {
      font-size: 13px;
      font-weight: 600;
      color: var(--cu-accent);
      margin-top: 2px;
    }
    /* Reward list */
    .reward-list {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }
    .reward-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      padding: 10px 0;
      border-bottom: 1px solid var(--cu-border);
    }
    .reward-row:last-child {
      border-bottom: none;
    }
    .reward-info {
      display: flex;
      flex-direction: column;
      gap: 2px;
    }
    .reward-name {
      font-size: 14px;
      font-weight: 600;
      color: var(--cu-ink);
    }
    .reward-cost {
      font-size: 12px;
      color: var(--cu-text-muted);
    }
    .reward-status {
      flex-shrink: 0;
    }
    /* Table */
    .table-wrap {
      overflow-x: auto;
    }
    .data-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 14px;
    }
    .data-table th,
    .data-table td {
      text-align: left;
      padding: 10px 12px;
      border-bottom: 1px solid var(--cu-border);
      color: var(--cu-ink);
    }
    .data-table th {
      font-weight: 600;
      font-size: 12px;
      color: var(--cu-text-muted);
      text-transform: uppercase;
      letter-spacing: 0.04em;
      background: var(--cu-bg);
    }
    .data-table tr:last-child td {
      border-bottom: none;
    }
    .points-cell {
      font-weight: 600;
      color: var(--cu-accent);
    }
    .points-value {
      font-weight: 700;
    }
    .points-positive {
      color: var(--cu-success);
    }
    .points-negative {
      color: var(--cu-danger);
    }
    /* Alerts & empty states */
    .cu-alert {
      padding: 12px 16px;
      border-radius: var(--cu-radius);
      font-size: 14px;
    }
    .cu-alert--error {
      background: var(--cu-danger-soft);
      color: var(--cu-danger);
    }
    .empty-text {
      font-size: 14px;
      color: var(--cu-text-muted);
      margin: 0;
    }
  `,
})
export class ProfilePage {
  protected readonly store = inject(AuthStore);
  private readonly rewardApi = inject(RewardApi);
  private readonly meApi = inject(MeApi);
  private readonly userApi = inject(UserApi);
  private readonly notif = inject(NotificationService);

  protected readonly ROLE_LABELS = ROLE_LABELS;
  protected readonly TRANSACTION_TYPE_LABELS = TRANSACTION_TYPE_LABELS;
  protected readonly formatDate = formatDate;
  protected readonly formatDateTime = formatDateTime;

  // Rewards section
  protected readonly rewardsLoading = signal(true);
  protected readonly rewardsError = signal<string | null>(null);
  protected readonly myRewards = signal<UserRewardList>({ user_id: 0, rewards: [] });

  // Event history section
  protected readonly eventsLoading = signal(true);
  protected readonly eventsError = signal<string | null>(null);
  protected readonly eventHistory = signal<UserEventHistoryList>({ user_id: 0, events: [], count: 0 });

  // Skill point transaction history section
  protected readonly txLoading = signal(true);
  protected readonly txError = signal<string | null>(null);
  protected readonly txHistory = signal<SkillPointHistory>({ user_id: 0, transactions: [], count: 0 });

  constructor() {
    this.loadMyRewards();
    this.loadEventHistory();
    this.loadTxHistory();
  }

  protected getInitials(user: { first_name: string; last_name: string }): string {
    return initials(user);
  }

  protected getFullName(user: { first_name: string; last_name: string; middle_name: string }): string {
    return fullName(user);
  }

  protected getEventStatusLabel(status: string): string {
    return EVENT_STATUS_LABELS[status as EventStatus] ?? status;
  }

  protected getEventStatusTone(status: string): 'success' | 'info' | 'warning' | 'neutral' | 'danger' {
    return EVENT_STATUS_TONE[status as EventStatus] ?? 'neutral';
  }

  private loadMyRewards(): void {
    this.rewardsLoading.set(true);
    this.rewardsError.set(null);
    this.rewardApi.myRewards().subscribe({
      next: (res) => {
        this.myRewards.set(res);
        this.rewardsLoading.set(false);
      },
      error: (e: ApiError) => {
        this.rewardsError.set(e.message);
        this.rewardsLoading.set(false);
      },
    });
  }

  private loadEventHistory(): void {
    this.eventsLoading.set(true);
    this.eventsError.set(null);
    this.meApi.events().subscribe({
      next: (res) => {
        this.eventHistory.set(res);
        this.eventsLoading.set(false);
      },
      error: (e: ApiError) => {
        this.eventsError.set(e.message);
        this.eventsLoading.set(false);
      },
    });
  }

  private loadTxHistory(): void {
    const user = this.store.user();
    if (!user) {
      this.txLoading.set(false);
      return;
    }
    this.txLoading.set(true);
    this.txError.set(null);
    this.userApi.skillPointHistory(user.id).subscribe({
      next: (res) => {
        this.txHistory.set(res);
        this.txLoading.set(false);
      },
      error: (e: ApiError) => {
        this.txError.set(e.message);
        this.txLoading.set(false);
      },
    });
  }

  protected claimReward(x: UserReward): void {
    this.rewardApi.claim(x.reward_id).subscribe({
      next: (res) => {
        this.notif.success(res.message);
        this.store.refresh().subscribe();
        this.loadMyRewards();
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
      },
    });
  }
}
