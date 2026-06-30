import { Component, inject } from '@angular/core';
import { NotificationService } from './notification.service';

/** Renders active toasts. Mount once near the app root. */
@Component({
  selector: 'cu-toast-container',
  template: `
    <div class="cu-toasts" aria-live="polite">
      @for (toast of notifications.toasts(); track toast.id) {
        <div class="cu-toast cu-toast--{{ toast.tone }}" (click)="notifications.dismiss(toast.id)">
          {{ toast.text }}
        </div>
      }
    </div>
  `,
  styles: `
    .cu-toasts {
      position: fixed;
      bottom: 24px;
      right: 24px;
      display: flex;
      flex-direction: column;
      gap: 10px;
      z-index: 1100;
      max-width: min(360px, calc(100vw - 48px));
    }
    .cu-toast {
      padding: 13px 16px;
      border-radius: var(--cu-radius);
      background: var(--cu-ink);
      color: #fff;
      font-size: 14px;
      font-weight: 500;
      box-shadow: var(--cu-shadow-lg);
      cursor: pointer;
      animation: cu-toast-in 0.18s ease;
    }
    .cu-toast--success {
      background: var(--cu-success);
    }
    .cu-toast--error {
      background: var(--cu-danger);
    }
    .cu-toast--info {
      background: var(--cu-accent);
    }
    @keyframes cu-toast-in {
      from {
        opacity: 0;
        transform: translateX(16px);
      }
    }
  `,
})
export class ToastContainer {
  protected readonly notifications = inject(NotificationService);
}
