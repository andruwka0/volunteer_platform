import { Component, input } from '@angular/core';

/** Big page title block inside a white panel (like "Мои курсы"). */
@Component({
  selector: 'cu-page-header',
  template: `
    <div class="cu-page-header">
      <div class="cu-page-header__text">
        <h1 class="cu-page-header__title">{{ title() }}</h1>
        @if (subtitle()) {
          <p class="cu-page-header__subtitle">{{ subtitle() }}</p>
        }
      </div>
      <div class="cu-page-header__actions">
        <ng-content />
      </div>
    </div>
  `,
  styles: `
    .cu-page-header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 16px;
      flex-wrap: wrap;
    }
    .cu-page-header__title {
      font-size: 36px;
      line-height: 1.1;
    }
    .cu-page-header__subtitle {
      margin: 8px 0 0;
      color: var(--cu-text-secondary);
      font-size: 15px;
    }
    .cu-page-header__actions:empty {
      display: none;
    }
    @media (max-width: 640px) {
      .cu-page-header__title {
        font-size: 28px;
      }
    }
  `,
})
export class PageHeader {
  readonly title = input.required<string>();
  readonly subtitle = input('');
}
