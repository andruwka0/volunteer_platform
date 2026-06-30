import { Component, input } from '@angular/core';

/** Empty/placeholder block. Project action buttons as content. */
@Component({
  selector: 'cu-empty-state',
  template: `
    <div class="cu-empty">
      <div class="cu-empty__icon" aria-hidden="true">{{ icon() }}</div>
      <h3 class="cu-empty__title">{{ title() }}</h3>
      @if (text()) {
        <p class="cu-empty__text">{{ text() }}</p>
      }
      <div class="cu-empty__actions">
        <ng-content />
      </div>
    </div>
  `,
  styles: `
    .cu-empty {
      display: flex;
      flex-direction: column;
      align-items: center;
      text-align: center;
      padding: 56px 24px;
      color: var(--cu-text-secondary);
    }
    .cu-empty__icon {
      font-size: 40px;
      margin-bottom: 12px;
      opacity: 0.7;
    }
    .cu-empty__title {
      font-size: 17px;
      color: var(--cu-text);
      margin-bottom: 6px;
    }
    .cu-empty__text {
      margin: 0 0 16px;
      max-width: 420px;
    }
    .cu-empty__actions:empty {
      display: none;
    }
  `,
})
export class EmptyState {
  readonly icon = input('📭');
  readonly title = input.required<string>();
  readonly text = input('');
}
