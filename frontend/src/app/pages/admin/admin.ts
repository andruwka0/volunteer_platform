import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiError } from '../../shared/api/api-error';
import { AuthStore } from '../../shared/auth/auth.store';
import {
  Badge,
  Button,
  Card,
  EmptyState,
  Modal,
  NotificationService,
  PageHeader,
  PillTab,
  PillTabs,
  Spinner,
} from '../../shared/ui';
import { RewardApi } from '../../entities/reward/reward.api';
import { CreateRewardPayload, Reward, UserReward } from '../../entities/reward/reward.model';
import { UserApi } from '../../entities/user/user.api';
import {
  ROLE_LABELS,
  User,
  UserRole,
  fullName,
} from '../../entities/user/user.model';

@Component({
  selector: 'cu-admin-page',
  imports: [
    ReactiveFormsModule,
    PageHeader,
    PillTabs,
    Button,
    Card,
    Badge,
    Spinner,
    EmptyState,
    Modal,
  ],
  template: `
    <div class="adm-page">
      <cu-page-header title="Администрирование" />

      <div class="adm-page__tabs">
        <cu-pill-tabs [tabs]="tabs" [(value)]="tab" />
      </div>

      <!-- ─── A) Пользователи ─── -->
      @if (tab() === 'users') {
        <section class="adm-section">
          <h2 class="adm-section__title">Пользователи</h2>

          <!-- Search form -->
          <cu-card>
            <form [formGroup]="searchForm" (ngSubmit)="searchUsers()" class="cu-form">
              <div class="cu-fields-row">
                <div class="cu-field">
                  <label class="cu-label" for="s-last">Фамилия *</label>
                  <input id="s-last" class="cu-input" type="text" formControlName="last_name"
                    placeholder="Иванов" />
                  @if (searchForm.get('last_name')?.invalid && searchForm.get('last_name')?.touched) {
                    <span class="cu-field__error">Обязательное поле</span>
                  }
                </div>
                <div class="cu-field">
                  <label class="cu-label" for="s-first">Имя *</label>
                  <input id="s-first" class="cu-input" type="text" formControlName="first_name"
                    placeholder="Иван" />
                  @if (searchForm.get('first_name')?.invalid && searchForm.get('first_name')?.touched) {
                    <span class="cu-field__error">Обязательное поле</span>
                  }
                </div>
                <div class="cu-field">
                  <label class="cu-label" for="s-mid">Отчество</label>
                  <input id="s-mid" class="cu-input" type="text" formControlName="middle_name"
                    placeholder="Иванович" />
                </div>
              </div>
              <div class="cu-form__actions">
                <cu-button type="submit" [variant]="'primary'"
                  [loading]="searchLoading()"
                  [disabled]="searchForm.invalid || searchLoading()">
                  Найти
                </cu-button>
              </div>
            </form>
          </cu-card>

          <!-- Search results -->
          @if (searchLoading()) {
            <div class="adm-center"><cu-spinner label="Поиск…" /></div>
          } @else if (searchError()) {
            <div class="cu-alert cu-alert--error">{{ searchError() }}</div>
          } @else if (searchRan() && users().length === 0) {
            <cu-empty-state icon="👤" title="Пользователи не найдены"
              text="Попробуйте изменить запрос" />
          } @else {
            <div class="adm-user-list">
              @for (u of users(); track u.id) {
                <cu-card>
                  <div class="adm-user__header">
                    <div>
                      <p class="adm-user__name">{{ fullName(u) }}</p>
                      <p class="adm-user__login">&#64;{{ u.login }}</p>
                    </div>
                    <div class="adm-user__badges">
                      <cu-badge [tone]="roleTone(u.role)">{{ ROLE_LABELS[u.role] }}</cu-badge>
                      <cu-badge [tone]="'info'">★ {{ u.skill_points }} СП</cu-badge>
                    </div>
                  </div>

                  <div class="adm-user__actions">
                    <!-- Promote -->
                    <div class="adm-action-row">
                      <label class="cu-label adm-action-row__label">Назначить роль:</label>
                      <select class="cu-select adm-action-row__select"
                        [id]="'promote-' + u.id"
                        (change)="setPromoteRole(u.id, $event)">
                        <option value="Organizer">Организатор</option>
                        <option value="Admin">Администратор</option>
                      </select>
                      <cu-button [size]="'sm'" [variant]="'outline'"
                        [loading]="promoteLoading() === u.id"
                        [disabled]="!!promoteLoading()"
                        (click)="promoteUser(u)">
                        Назначить
                      </cu-button>
                    </div>

                    <!-- Award -->
                    <div class="adm-action-row">
                      <label class="cu-label adm-action-row__label">Начислить баллы:</label>
                      <input class="cu-input adm-action-row__input-num" type="number" min="1"
                        placeholder="Кол-во"
                        [id]="'award-pts-' + u.id"
                        (input)="setAwardPoints(u.id, $event)" />
                      <input class="cu-input adm-action-row__input-reason" type="text"
                        placeholder="Причина"
                        [id]="'award-reason-' + u.id"
                        (input)="setAwardReason(u.id, $event)" />
                      <cu-button [size]="'sm'" [variant]="'subtle'"
                        [loading]="awardLoading() === u.id"
                        [disabled]="!!awardLoading()"
                        (click)="awardUser(u)">
                        Начислить
                      </cu-button>
                    </div>

                    <!-- Rewards modal button -->
                    <div class="adm-action-row">
                      <cu-button [size]="'sm'" [variant]="'ghost'" (click)="openRewardsModal(u)">
                        Награды пользователя
                      </cu-button>
                    </div>
                  </div>
                </cu-card>
              }
            </div>
          }
        </section>

        <!-- User rewards modal -->
        @if (modalUser() !== null) {
          <cu-modal
            [title]="'Награды: ' + fullName(modalUser()!)"
            (closed)="closeModal()">
            @if (modalLoading()) {
              <div class="adm-center"><cu-spinner label="Загрузка наград…" /></div>
            } @else if (modalError()) {
              <div class="cu-alert cu-alert--error">{{ modalError() }}</div>
            } @else if (modalRewards().length === 0) {
              <cu-empty-state icon="🎁" title="Нет наград" text="У пользователя нет наград" />
            } @else {
              <div class="adm-rewards-list">
                @for (r of modalRewards(); track r.reward_id) {
                  <div class="adm-reward-row">
                    <div class="adm-reward-row__info">
                      <p class="adm-reward-row__name">{{ r.name }}</p>
                      <p class="adm-reward-row__desc">{{ r.description }}</p>
                      <p class="adm-reward-row__cost">★ {{ r.cost }} СП</p>
                    </div>
                    <div class="adm-reward-row__status">
                      @if (r.picked_up) {
                        <cu-badge [tone]="'success'">Выдано</cu-badge>
                      } @else if (r.claimed && !r.picked_up) {
                        <cu-button [size]="'sm'"
                          [loading]="confirmLoading() === r.reward_id"
                          [disabled]="!!confirmLoading()"
                          (click)="confirmPickup(modalUser()!.id, r.reward_id)">
                          Подтвердить выдачу
                        </cu-button>
                      } @else {
                        <cu-badge [tone]="'neutral'">Не заявлено</cu-badge>
                      }
                    </div>
                  </div>
                }
              </div>
            }
          </cu-modal>
        }
      }

      <!-- ─── B) Награды ─── -->
      @if (tab() === 'rewards') {
        <section class="adm-section">
          <h2 class="adm-section__title">Награды</h2>

          @if (rewardsLoading()) {
            <div class="adm-center"><cu-spinner label="Загрузка наград…" /></div>
          } @else if (rewardsError()) {
            <div class="cu-alert cu-alert--error">{{ rewardsError() }}</div>
          } @else if (rewards().length === 0) {
            <cu-empty-state icon="🎁" title="Нет наград" text="Создайте первую награду ниже" />
          } @else {
            <div class="adm-grid">
              @for (r of rewards(); track r.id) {
                <cu-card>
                  <p class="adm-card__title">{{ r.name }}</p>
                  @if (r.description) {
                    <p class="adm-card__meta">{{ r.description }}</p>
                  }
                  <p class="adm-card__meta">★ {{ r.cost }} СП</p>
                </cu-card>
              }
            </div>
          }

          <h3 class="adm-section__subtitle">Создать награду</h3>
          <cu-card>
            <form [formGroup]="rewardForm" (ngSubmit)="submitReward()" class="cu-form">
              <div class="cu-field">
                <label class="cu-label" for="rw-name">Название *</label>
                <input id="rw-name" class="cu-input" type="text" formControlName="name"
                  placeholder="Название награды" />
                @if (rewardForm.get('name')?.invalid && rewardForm.get('name')?.touched) {
                  <span class="cu-field__error">Обязательное поле</span>
                }
              </div>
              <div class="cu-field">
                <label class="cu-label" for="rw-desc">Описание</label>
                <textarea id="rw-desc" class="cu-textarea" formControlName="description"
                  rows="3"></textarea>
              </div>
              <div class="cu-field">
                <label class="cu-label" for="rw-img">URL изображения</label>
                <input id="rw-img" class="cu-input" type="text" formControlName="image_url"
                  placeholder="https://…" />
              </div>
              <div class="cu-field">
                <label class="cu-label" for="rw-cost">Стоимость (СП) *</label>
                <input id="rw-cost" class="cu-input" type="number" min="1"
                  formControlName="cost" />
                @if (rewardForm.get('cost')?.invalid && rewardForm.get('cost')?.touched) {
                  <span class="cu-field__error">Укажите стоимость</span>
                }
              </div>
              <div class="cu-form__actions">
                <cu-button type="submit" [variant]="'primary'"
                  [loading]="rewardSubmitting()"
                  [disabled]="rewardForm.invalid || rewardSubmitting()">
                  Создать награду
                </cu-button>
              </div>
            </form>
          </cu-card>
        </section>
      }
    </div>
  `,
  styles: `
    .adm-page {
      max-width: 1100px;
      margin: 0 auto;
      padding: 24px 16px;
    }
    .adm-page__tabs {
      margin: 16px 0 24px;
    }
    .adm-section {
      display: flex;
      flex-direction: column;
      gap: 20px;
    }
    .adm-section__title {
      font-size: 20px;
      font-weight: 700;
      color: var(--cu-text);
      margin: 0;
    }
    .adm-section__subtitle {
      font-size: 17px;
      font-weight: 600;
      color: var(--cu-text);
      margin: 8px 0 0;
    }
    .adm-center {
      display: flex;
      justify-content: center;
      padding: 32px 0;
    }
    .adm-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
      gap: 16px;
    }
    .adm-card__title {
      font-weight: 700;
      font-size: 15px;
      margin: 0 0 6px;
      color: var(--cu-text);
    }
    .adm-card__meta {
      font-size: 13px;
      color: var(--cu-text-secondary);
      margin: 2px 0;
    }
    .adm-user-list {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
    .adm-user__header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 12px;
      margin-bottom: 12px;
    }
    .adm-user__name {
      font-weight: 700;
      font-size: 15px;
      margin: 0 0 2px;
      color: var(--cu-text);
    }
    .adm-user__login {
      font-size: 13px;
      color: var(--cu-text-secondary);
      margin: 0;
    }
    .adm-user__badges {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
      flex-shrink: 0;
    }
    .adm-user__actions {
      display: flex;
      flex-direction: column;
      gap: 10px;
      border-top: 1px solid var(--cu-border);
      padding-top: 12px;
    }
    .adm-action-row {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 8px;
    }
    .adm-action-row__label {
      margin: 0;
      flex-shrink: 0;
      font-size: 13px;
    }
    .adm-action-row__select {
      width: 160px;
      flex-shrink: 0;
    }
    .adm-action-row__input-num {
      width: 90px;
      flex-shrink: 0;
    }
    .adm-action-row__input-reason {
      flex: 1;
      min-width: 120px;
    }
    .adm-rewards-list {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }
    .adm-reward-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      padding: 12px;
      border: 1px solid var(--cu-border);
      border-radius: var(--cu-radius-md);
    }
    .adm-reward-row__info {
      flex: 1;
    }
    .adm-reward-row__name {
      font-weight: 600;
      font-size: 14px;
      margin: 0 0 2px;
      color: var(--cu-text);
    }
    .adm-reward-row__desc {
      font-size: 13px;
      color: var(--cu-text-secondary);
      margin: 0 0 2px;
    }
    .adm-reward-row__cost {
      font-size: 13px;
      color: var(--cu-text-secondary);
      margin: 0;
    }
    .adm-reward-row__status {
      flex-shrink: 0;
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
      min-width: 140px;
    }
    .cu-form__actions {
      display: flex;
      justify-content: flex-end;
      padding-top: 4px;
    }
  `,
})
export class AdminPage {
  private readonly userApi = inject(UserApi);
  private readonly rewardApi = inject(RewardApi);
  private readonly notif = inject(NotificationService);
  private readonly fb = inject(FormBuilder);

  protected readonly store = inject(AuthStore);

  protected readonly fullName = fullName;
  protected readonly ROLE_LABELS = ROLE_LABELS;

  protected readonly tab = signal('users');
  protected readonly tabs: PillTab[] = [
    { value: 'users', label: 'Пользователи' },
    { value: 'rewards', label: 'Награды' },
  ];

  // ── Users / Search ──────────────────────────────────────

  protected readonly users = signal<User[]>([]);
  protected readonly searchLoading = signal(false);
  protected readonly searchError = signal<string | null>(null);
  protected readonly searchRan = signal(false);

  // Per-user action state: stores userId for pending promote/award
  protected readonly promoteLoading = signal<number | null>(null);
  protected readonly awardLoading = signal<number | null>(null);

  // Per-user promote target role (keyed by userId)
  private readonly promoteRoles = new Map<number, UserRole>();

  // Per-user award fields (keyed by userId)
  private readonly awardPointsMap = new Map<number, number>();
  private readonly awardReasonMap = new Map<number, string>();

  // Last search params (for re-running after actions)
  private lastSearch: { first: string; last: string; middle?: string } | null = null;

  // ── Modal ───────────────────────────────────────────────

  protected readonly modalUser = signal<User | null>(null);
  protected readonly modalRewards = signal<UserReward[]>([]);
  protected readonly modalLoading = signal(false);
  protected readonly modalError = signal<string | null>(null);
  protected readonly confirmLoading = signal<number | null>(null);

  // ── Rewards catalog ─────────────────────────────────────

  protected readonly rewards = signal<Reward[]>([]);
  protected readonly rewardsLoading = signal(true);
  protected readonly rewardsError = signal<string | null>(null);
  protected readonly rewardSubmitting = signal(false);

  // ── Forms ───────────────────────────────────────────────

  protected readonly searchForm = this.fb.group({
    first_name: ['', Validators.required],
    last_name: ['', Validators.required],
    middle_name: [''],
  });

  protected readonly rewardForm = this.fb.group({
    name: ['', Validators.required],
    description: [''],
    image_url: [''],
    cost: [null as number | null, [Validators.required, Validators.min(1)]],
  });

  // ── Lifecycle ───────────────────────────────────────────

  constructor() {
    this.loadRewards();
  }

  private loadRewards(): void {
    this.rewardsLoading.set(true);
    this.rewardsError.set(null);
    this.rewardApi.list().subscribe({
      next: (res) => {
        this.rewards.set(res);
        this.rewardsLoading.set(false);
      },
      error: (e: ApiError) => {
        this.rewardsError.set(e.message);
        this.rewardsLoading.set(false);
      },
    });
  }

  // ── Search ──────────────────────────────────────────────

  protected searchUsers(): void {
    if (this.searchForm.invalid) {
      this.searchForm.markAllAsTouched();
      return;
    }
    const v = this.searchForm.getRawValue();
    const first = v.first_name ?? '';
    const last = v.last_name ?? '';
    const middle = v.middle_name ?? undefined;
    this.lastSearch = { first, last, middle };
    this.runSearch(first, last, middle);
  }

  private runSearch(first: string, last: string, middle?: string): void {
    this.searchLoading.set(true);
    this.searchError.set(null);
    this.searchRan.set(false);
    this.userApi.search(first, last, middle).subscribe({
      next: (res) => {
        this.users.set(res.users);
        this.searchLoading.set(false);
        this.searchRan.set(true);
      },
      error: (e: ApiError) => {
        this.searchError.set(e.message);
        this.searchLoading.set(false);
        this.searchRan.set(true);
      },
    });
  }

  // ── Promote ─────────────────────────────────────────────

  protected setPromoteRole(userId: number, event: Event): void {
    const sel = event.target as HTMLSelectElement;
    this.promoteRoles.set(userId, sel.value as UserRole);
  }

  protected promoteUser(u: User): void {
    const role = this.promoteRoles.get(u.id) ?? 'Organizer';
    this.promoteLoading.set(u.id);
    this.userApi.promote(u.id, role).subscribe({
      next: () => {
        this.notif.success(`${fullName(u)} назначен: ${ROLE_LABELS[role]}`);
        this.promoteLoading.set(null);
        this.rerunLastSearch();
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
        this.promoteLoading.set(null);
      },
    });
  }

  // ── Award ───────────────────────────────────────────────

  protected setAwardPoints(userId: number, event: Event): void {
    const inp = event.target as HTMLInputElement;
    this.awardPointsMap.set(userId, Number(inp.value));
  }

  protected setAwardReason(userId: number, event: Event): void {
    const inp = event.target as HTMLInputElement;
    this.awardReasonMap.set(userId, inp.value);
  }

  protected awardUser(u: User): void {
    const points = this.awardPointsMap.get(u.id) ?? 0;
    const reason = this.awardReasonMap.get(u.id) ?? '';
    if (!points || points <= 0) {
      this.notif.error('Укажите количество баллов');
      return;
    }

    this.awardLoading.set(u.id);
    this.userApi.award(u.id, points, reason).subscribe({
      next: () => {
        this.notif.success(`Начислено ${points} СП для ${fullName(u)}`);
        this.awardLoading.set(null);
        this.awardPointsMap.delete(u.id);
        this.awardReasonMap.delete(u.id);
        // If awarding current user — refresh auth store
        if (u.id === this.store.user()?.id) {
          this.store.refresh().subscribe();
        }
        this.rerunLastSearch();
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
        this.awardLoading.set(null);
      },
    });
  }

  private rerunLastSearch(): void {
    if (this.lastSearch) {
      const { first, last, middle } = this.lastSearch;
      this.runSearch(first, last, middle);
    }
  }

  // ── Role tone helper ────────────────────────────────────

  protected roleTone(role: UserRole): 'success' | 'info' | 'warning' | 'danger' | 'neutral' | 'purple' {
    if (role === 'Admin') return 'danger';
    if (role === 'Organizer') return 'purple';
    return 'neutral';
  }

  // ── Modal ───────────────────────────────────────────────

  protected openRewardsModal(u: User): void {
    this.modalUser.set(u);
    this.loadModalRewards(u.id);
  }

  protected closeModal(): void {
    this.modalUser.set(null);
    this.modalRewards.set([]);
    this.modalError.set(null);
  }

  private loadModalRewards(userId: number): void {
    this.modalLoading.set(true);
    this.modalError.set(null);
    this.userApi.rewardsByAdmin(userId).subscribe({
      next: (res) => {
        this.modalRewards.set(res.rewards);
        this.modalLoading.set(false);
      },
      error: (e: ApiError) => {
        this.modalError.set(e.message);
        this.modalLoading.set(false);
      },
    });
  }

  protected confirmPickup(userId: number, rewardId: number): void {
    this.confirmLoading.set(rewardId);
    this.rewardApi.confirmPickup(userId, rewardId).subscribe({
      next: () => {
        this.notif.success('Выдача подтверждена');
        this.confirmLoading.set(null);
        this.loadModalRewards(userId);
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
        this.confirmLoading.set(null);
      },
    });
  }

  // ── Create Reward ───────────────────────────────────────

  protected submitReward(): void {
    if (this.rewardForm.invalid) {
      this.rewardForm.markAllAsTouched();
      return;
    }
    const v = this.rewardForm.getRawValue();
    const payload: CreateRewardPayload = {
      name: v.name ?? '',
      description: v.description ?? '',
      image_url: v.image_url ?? '',
      cost: v.cost ?? 0,
    };

    this.rewardSubmitting.set(true);
    this.rewardApi.create(payload).subscribe({
      next: () => {
        this.notif.success('Награда создана');
        this.rewardForm.reset({ name: '', description: '', image_url: '', cost: null });
        this.rewardSubmitting.set(false);
        this.loadRewards();
      },
      error: (e: ApiError) => {
        this.notif.error(e.message);
        this.rewardSubmitting.set(false);
      },
    });
  }
}
