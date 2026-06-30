import { Component, input } from '@angular/core';

export type BadgeTone = 'success' | 'info' | 'warning' | 'danger' | 'neutral' | 'purple';

/** Small status pill. Usage: <cu-badge tone="success">Текст</cu-badge> */
@Component({
  selector: 'cu-badge',
  template: `<span [class]="'cu-badge cu-badge--' + tone()"><ng-content /></span>`,
  styles: `
    :host {
      display: inline-flex;
    }
    .cu-badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 3px 10px;
      border-radius: var(--cu-radius-pill);
      font-size: 12px;
      font-weight: 600;
      line-height: 1.4;
      white-space: nowrap;
    }
    .cu-badge--success {
      background: var(--cu-success-soft);
      color: var(--cu-success);
    }
    .cu-badge--info {
      background: var(--cu-info-soft);
      color: var(--cu-info);
    }
    .cu-badge--warning {
      background: var(--cu-warning-soft);
      color: var(--cu-warning);
    }
    .cu-badge--danger {
      background: var(--cu-danger-soft);
      color: var(--cu-danger);
    }
    .cu-badge--neutral {
      background: var(--cu-neutral-soft);
      color: var(--cu-neutral);
    }
    .cu-badge--purple {
      background: var(--cu-purple-soft);
      color: var(--cu-purple);
    }
  `,
})
export class Badge {
  readonly tone = input<BadgeTone>('neutral');
}
