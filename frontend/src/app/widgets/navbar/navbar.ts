import { Component, computed, inject } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { AuthStore } from '../../shared/auth/auth.store';
import { Avatar } from '../../shared/ui';
import { fullName, initials, ROLE_LABELS } from '../../entities/user/user.model';

interface NavItem {
  label: string;
  link: string;
}

/** Top navigation bar — Central University style. */
@Component({
  selector: 'cu-navbar',
  imports: [RouterLink, RouterLinkActive, Avatar],
  template: `
    <header class="nav">
      <div class="nav__inner cu-container">
        <nav class="nav__links" aria-label="Основная навигация">
          @for (item of links(); track item.link) {
            <a
              class="nav__link"
              [routerLink]="item.link"
              routerLinkActive="nav__link--active"
              >{{ item.label }}</a
            >
          }
        </nav>

        <div class="nav__right">
          <span class="nav__chip" title="Ваши Skill Points">
            {{ store.skillPoints() }} SP
          </span>

          @if (user(); as u) {
            <div class="nav__user">
              <div class="nav__user-meta">
                <span class="nav__user-name">{{ name() }}</span>
                <span class="nav__user-role">{{ roleLabel() }}</span>
              </div>
              <cu-avatar [initials]="userInitials()" [size]="38" />
              <button type="button" class="nav__logout" (click)="store.logout()" title="Выйти">
                ⎋
              </button>
            </div>
          }
        </div>
      </div>
    </header>
  `,
  styles: `
    .nav {
      background: var(--cu-surface);
      border-bottom: 1px solid var(--cu-border);
      position: sticky;
      top: 0;
      z-index: 100;
    }
    .nav__inner {
      height: var(--cu-navbar-h);
      display: flex;
      align-items: center;
      gap: 28px;
    }
    .nav__brand {
      display: flex;
      align-items: center;
      gap: 10px;
      color: var(--cu-text);
      flex-shrink: 0;
    }
    .nav__logo {
      width: 34px;
      height: 34px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      background: var(--cu-ink);
      color: #fff;
      border-radius: 9px;
      font-size: 16px;
    }
    .nav__brand-text {
      display: flex;
      flex-direction: column;
      line-height: 1.05;
    }
    .nav__brand-top {
      font-size: 13px;
      font-weight: 800;
      letter-spacing: 0.02em;
    }
    .nav__brand-bottom {
      font-size: 11px;
      font-weight: 700;
      color: var(--cu-text-muted);
      letter-spacing: 0.06em;
    }
    .nav__links {
      display: flex;
      align-items: center;
      gap: 4px;
      flex: 1;
    }
    .nav__link {
      color: var(--cu-text-secondary);
      font-weight: 600;
      font-size: 14px;
      padding: 8px 14px;
      border-radius: var(--cu-radius-pill);
      white-space: nowrap;
    }
    .nav__link:hover {
      background: var(--cu-surface-3);
      color: var(--cu-text);
    }
    .nav__link--active {
      color: var(--cu-accent);
      background: var(--cu-accent-soft);
    }
    .nav__right {
      display: flex;
      align-items: center;
      gap: 14px;
      flex-shrink: 0;
    }
    .nav__chip {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      background: var(--cu-purple-soft);
      color: var(--cu-purple);
      font-weight: 700;
      font-size: 13px;
      padding: 7px 14px;
      border-radius: var(--cu-radius-pill);
    }
    .nav__chip-icon {
      font-size: 13px;
    }
    .nav__user {
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .nav__user-meta {
      display: flex;
      flex-direction: column;
      align-items: flex-end;
      line-height: 1.15;
    }
    .nav__user-name {
      font-size: 13px;
      font-weight: 700;
    }
    .nav__user-role {
      font-size: 11px;
      color: var(--cu-text-muted);
    }
    .nav__logout {
      border: none;
      background: var(--cu-surface-3);
      width: 34px;
      height: 34px;
      border-radius: 50%;
      cursor: pointer;
      color: var(--cu-text-secondary);
      font-size: 15px;
    }
    .nav__logout:hover {
      background: var(--cu-border-strong);
      color: var(--cu-text);
    }
    @media (max-width: 860px) {
      .nav__links {
        display: none;
      }
      .nav__user-meta {
        display: none;
      }
    }
  `,
})
export class Navbar {
  protected readonly store = inject(AuthStore);
  protected readonly user = this.store.user;

  protected readonly links = computed<NavItem[]>(() => {
    const items: NavItem[] = [
      { label: 'Мероприятия', link: '/events' },
      { label: 'Награды', link: '/rewards' },
      { label: 'Профиль', link: '/profile' },
    ];
    if (this.store.isOrganizer()) {
      items.push({ label: 'Организатору', link: '/organizer' });
    }
    if (this.store.isAdmin()) {
      items.push({ label: 'Админка', link: '/admin' });
    }
    return items;
  });

  protected readonly name = computed(() => {
    const u = this.user();
    return u ? fullName(u) || u.login : '';
  });
  protected readonly userInitials = computed(() => {
    const u = this.user();
    return u ? initials(u) || u.login.slice(0, 2).toUpperCase() : '?';
  });
  protected readonly roleLabel = computed(() => {
    const u = this.user();
    return u ? ROLE_LABELS[u.role] : '';
  });
}
