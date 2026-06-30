import { Component, input, model } from '@angular/core';

export interface PillTab {
  value: string;
  label: string;
}

/**
 * Filter pills like the Central University "Мои курсы" tabs.
 * Active tab is dark-filled. Two-way bind the selected value.
 * Usage: <cu-pill-tabs [tabs]="tabs" [(value)]="active" />
 */
@Component({
  selector: 'cu-pill-tabs',
  template: `
    <div class="cu-pills" role="tablist">
      @for (tab of tabs(); track tab.value) {
        <button
          type="button"
          role="tab"
          [attr.aria-selected]="value() === tab.value"
          class="cu-pill"
          [class.cu-pill--active]="value() === tab.value"
          (click)="value.set(tab.value)"
        >
          {{ tab.label }}
        </button>
      }
    </div>
  `,
  styles: `
    .cu-pills {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }
    .cu-pill {
      border: none;
      background: transparent;
      color: var(--cu-text);
      padding: 9px 18px;
      border-radius: var(--cu-radius-pill);
      font-size: 14px;
      font-weight: 600;
      cursor: pointer;
      transition:
        background 0.15s ease,
        color 0.15s ease;
    }
    .cu-pill:hover:not(.cu-pill--active) {
      background: var(--cu-surface-3);
    }
    .cu-pill--active {
      background: var(--cu-ink);
      color: var(--cu-text-on-dark);
    }
  `,
})
export class PillTabs {
  readonly tabs = input.required<PillTab[]>();
  readonly value = model.required<string>();
}
