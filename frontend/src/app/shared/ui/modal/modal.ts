import { Component, input, output } from '@angular/core';

/**
 * Centered modal dialog. Parent controls visibility with @if and listens to (closed).
 * Usage:
 *   @if (open()) {
 *     <cu-modal title="Заголовок" (closed)="open.set(false)"> ...content... </cu-modal>
 *   }
 */
@Component({
  selector: 'cu-modal',
  template: `
    <div class="cu-modal__backdrop" (click)="closed.emit()">
      <div
        class="cu-modal"
        role="dialog"
        aria-modal="true"
        [style.max-width.px]="width()"
        (click)="$event.stopPropagation()"
      >
        <header class="cu-modal__head">
          <h2 class="cu-modal__title">{{ title() }}</h2>
          <button type="button" class="cu-modal__close" aria-label="Закрыть" (click)="closed.emit()">
            ✕
          </button>
        </header>
        <div class="cu-modal__body">
          <ng-content />
        </div>
      </div>
    </div>
  `,
  host: {
    '(document:keydown.escape)': 'closed.emit()',
  },
  styles: `
    .cu-modal__backdrop {
      position: fixed;
      inset: 0;
      background: rgba(15, 23, 42, 0.45);
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 24px;
      z-index: 1000;
      animation: cu-fade 0.15s ease;
    }
    .cu-modal {
      background: var(--cu-surface);
      border-radius: var(--cu-radius-lg);
      box-shadow: var(--cu-shadow-lg);
      width: 100%;
      max-height: 90vh;
      overflow: auto;
      animation: cu-pop 0.15s ease;
    }
    .cu-modal__head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 20px 24px;
      border-bottom: 1px solid var(--cu-border);
      position: sticky;
      top: 0;
      background: var(--cu-surface);
    }
    .cu-modal__title {
      font-size: 19px;
    }
    .cu-modal__close {
      border: none;
      background: var(--cu-surface-3);
      width: 32px;
      height: 32px;
      border-radius: 50%;
      cursor: pointer;
      color: var(--cu-text-secondary);
      font-size: 14px;
    }
    .cu-modal__close:hover {
      background: var(--cu-border-strong);
    }
    .cu-modal__body {
      padding: 24px;
    }
    @keyframes cu-fade {
      from {
        opacity: 0;
      }
    }
    @keyframes cu-pop {
      from {
        opacity: 0;
        transform: translateY(8px) scale(0.98);
      }
    }
  `,
})
export class Modal {
  readonly title = input('');
  readonly width = input(520);
  readonly closed = output<void>();
}
