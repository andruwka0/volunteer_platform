import { Component, input } from '@angular/core';

export type ButtonVariant = 'primary' | 'ghost' | 'outline' | 'danger' | 'purple' | 'subtle';
export type ButtonSize = 'sm' | 'md' | 'lg';

/**
 * Central University button.
 * Usage: <cu-button variant="primary" (click)="...">Текст</cu-button>
 */
@Component({
  selector: 'cu-button',
  template: `
    <button
      [type]="type()"
      [disabled]="disabled() || loading()"
      [class]="'cu-btn cu-btn--' + variant() + ' cu-btn--' + size()"
      [class.cu-btn--full]="full()"
    >
      @if (loading()) {
        <span class="cu-btn__spinner" aria-hidden="true"></span>
      }
      <ng-content />
    </button>
  `,
  styles: `
    :host {
      display: inline-flex;
    }
    :host(.cu-btn-host--full),
    :host([full]) {
      width: 100%;
    }
    .cu-btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      border: 1px solid transparent;
      border-radius: var(--cu-radius-pill);
      font-weight: 600;
      cursor: pointer;
      white-space: nowrap;
      transition:
        background 0.15s ease,
        color 0.15s ease,
        border-color 0.15s ease,
        transform 0.05s ease;
    }
    .cu-btn:active {
      transform: translateY(1px);
    }
    .cu-btn:disabled {
      opacity: 0.55;
      cursor: not-allowed;
    }
    .cu-btn--full {
      width: 100%;
    }

    .cu-btn--sm {
      padding: 6px 14px;
      font-size: 13px;
    }
    .cu-btn--md {
      padding: 10px 20px;
      font-size: 14px;
    }
    .cu-btn--lg {
      padding: 13px 26px;
      font-size: 15px;
    }

    .cu-btn--primary {
      background: var(--cu-ink);
      color: var(--cu-text-on-dark);
    }
    .cu-btn--primary:hover:not(:disabled) {
      background: var(--cu-ink-hover);
    }

    .cu-btn--purple {
      background: var(--cu-purple);
      color: #fff;
    }
    .cu-btn--purple:hover:not(:disabled) {
      filter: brightness(0.95);
    }

    .cu-btn--danger {
      background: var(--cu-danger);
      color: #fff;
    }
    .cu-btn--danger:hover:not(:disabled) {
      filter: brightness(0.95);
    }

    .cu-btn--outline {
      background: var(--cu-surface);
      color: var(--cu-text);
      border-color: var(--cu-border-strong);
    }
    .cu-btn--outline:hover:not(:disabled) {
      background: var(--cu-surface-3);
    }

    .cu-btn--ghost {
      background: transparent;
      color: var(--cu-text);
    }
    .cu-btn--ghost:hover:not(:disabled) {
      background: var(--cu-surface-3);
    }

    .cu-btn--subtle {
      background: var(--cu-accent-soft);
      color: var(--cu-accent);
    }
    .cu-btn--subtle:hover:not(:disabled) {
      filter: brightness(0.97);
    }

    .cu-btn__spinner {
      width: 14px;
      height: 14px;
      border: 2px solid currentColor;
      border-right-color: transparent;
      border-radius: 50%;
      animation: cu-btn-spin 0.6s linear infinite;
    }
    @keyframes cu-btn-spin {
      to {
        transform: rotate(360deg);
      }
    }
  `,
})
export class Button {
  readonly variant = input<ButtonVariant>('primary');
  readonly size = input<ButtonSize>('md');
  readonly type = input<'button' | 'submit' | 'reset'>('button');
  readonly disabled = input(false);
  readonly loading = input(false);
  readonly full = input(false);
}
