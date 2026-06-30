import { Component, input } from '@angular/core';

/** Centered loading spinner. */
@Component({
  selector: 'cu-spinner',
  template: `
    <div class="cu-spinner-wrap">
      <span class="cu-spinner" [style.width.px]="size()" [style.height.px]="size()"></span>
      @if (label()) {
        <span class="cu-spinner-label">{{ label() }}</span>
      }
    </div>
  `,
  styles: `
    .cu-spinner-wrap {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 12px;
      padding: 32px;
      color: var(--cu-text-secondary);
    }
    .cu-spinner {
      border: 3px solid var(--cu-surface-3);
      border-top-color: var(--cu-accent);
      border-radius: 50%;
      animation: cu-spin 0.7s linear infinite;
    }
    .cu-spinner-label {
      font-size: 14px;
    }
    @keyframes cu-spin {
      to {
        transform: rotate(360deg);
      }
    }
  `,
})
export class Spinner {
  readonly size = input(32);
  readonly label = input('');
}
