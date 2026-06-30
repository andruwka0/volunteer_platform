import { Component, computed, inject, signal } from '@angular/core';
import { RewardApi } from '../../entities/reward/reward.api';
import { Reward, UserRewardList } from '../../entities/reward/reward.model';
import { ApiError } from '../../shared/api/api-error';
import { AuthStore } from '../../shared/auth/auth.store';
import { Badge, Button, PageHeader, Spinner } from '../../shared/ui';
import { NotificationService } from '../../shared/ui';
import { RewardCard } from '../../widgets/reward-card/reward-card';

@Component({
  selector: 'cu-rewards-page',
  imports: [PageHeader, Spinner, Badge, Button, RewardCard],
  template: `
    <cu-page-header
      title="Магазин наград"
      [subtitle]="'У вас ' + store.skillPoints() + ' SP'"
    />

    @if (loading()) {
      <cu-spinner label="Загружаем награды..." />
    } @else if (error()) {
      <div class="cu-alert cu-alert--error">{{ error() }}</div>
    } @else {
      <div class="rewards-grid">
        @for (r of rewards(); track r.id) {
          <cu-reward-card
            [name]="r.name"
            [description]="r.description"
            [cost]="r.cost"
            [imageUrl]="r.image_url"
          >
            @if (isClaimed(r.id)) {
              <cu-badge [tone]="'success'">Куплено</cu-badge>
            } @else {
              <cu-button
                [disabled]="store.skillPoints() < r.cost"
                [loading]="acting() === r.id"
                (click)="claim(r)"
              >
                Получить за {{ r.cost }} SP
              </cu-button>
            }
          </cu-reward-card>
        } @empty {
          <p class="empty-text">Нет доступных наград.</p>
        }
      </div>
    }
  `,
  styles: `
    :host {
      display: block;
      padding: 24px;
    }
    .rewards-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
      gap: 16px;
      margin-top: 24px;
    }
    .cu-alert {
      padding: 12px 16px;
      border-radius: var(--cu-radius);
      font-size: 14px;
      margin-top: 16px;
    }
    .cu-alert--error {
      background: var(--cu-danger-soft);
      color: var(--cu-danger);
    }
    .empty-text {
      color: var(--cu-text-muted);
      font-size: 14px;
      grid-column: 1 / -1;
    }
  `,
})
export class RewardsPage {
  protected readonly store = inject(AuthStore);
  private readonly rewardApi = inject(RewardApi);
  private readonly notif = inject(NotificationService);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly rewards = signal<Reward[]>([]);
  protected readonly myRewards = signal<UserRewardList>({ user_id: 0, rewards: [] });
  protected readonly acting = signal<number | null>(null);

  private readonly claimedIds = computed<Set<number>>(() => {
    const set = new Set<number>();
    for (const x of this.myRewards().rewards) {
      if (x.claimed) set.add(x.reward_id);
    }
    return set;
  });

  protected isClaimed(rewardId: number): boolean {
    return this.claimedIds().has(rewardId);
  }

  constructor() {
    this.loadAll();
  }

  private loadAll(): void {
    this.loading.set(true);
    this.error.set(null);

    let catalogDone = false;
    let myDone = false;
    let catalogErr: string | null = null;
    let myErr: string | null = null;

    const finish = () => {
      if (!catalogDone || !myDone) return;
      this.error.set(catalogErr ?? myErr ?? null);
      this.loading.set(false);
    };

    this.rewardApi.list().subscribe({
      next: (res) => {
        this.rewards.set(res);
        catalogDone = true;
        finish();
      },
      error: (e: ApiError) => {
        catalogErr = e.message;
        catalogDone = true;
        finish();
      },
    });

    this.rewardApi.myRewards().subscribe({
      next: (res) => {
        this.myRewards.set(res);
        myDone = true;
        finish();
      },
      error: (e: ApiError) => {
        myErr = e.message;
        myDone = true;
        finish();
      },
    });
  }

  private reloadMyRewards(): void {
    this.rewardApi.myRewards().subscribe({
      next: (res) => this.myRewards.set(res),
      error: () => {},
    });
  }

  protected claim(r: Reward): void {
    if (this.acting() !== null) return;
    this.acting.set(r.id);
    this.rewardApi.claim(r.id).subscribe({
      next: (res) => {
        this.notif.success(res.message);
        this.acting.set(null);
        this.store.refresh().subscribe();
        this.reloadMyRewards();
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
        this.acting.set(null);
      },
    });
  }
}
