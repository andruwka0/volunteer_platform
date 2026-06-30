import { Component, input, signal } from '@angular/core';
import { Card } from '../../shared/ui';

/**
 * Reusable presentational card for a single reward.
 * Project a footer (action/status) via <ng-content>.
 */
@Component({
  selector: 'cu-reward-card',
  imports: [Card],
  template: `
    <cu-card>
      <!-- Image area -->
      <div class="rc-image-wrap">
        @if (imageUrl() && !imgFailed()) {
          <img
            [src]="imageUrl()"
            [alt]="name()"
            class="rc-image"
            (error)="imgFailed.set(true)"
          />
        } @else {
          <div class="rc-placeholder" aria-hidden="true">
            <span class="rc-placeholder-letter">{{ name()[0]?.toUpperCase() }}</span>
          </div>
        }
      </div>

      <!-- Content -->
      <div class="rc-body">
        <p class="rc-name">{{ name() }}</p>
        @if (description()) {
          <p class="rc-description">{{ description() }}</p>
        }
        <p class="rc-cost">&#9733; {{ cost() }} SP</p>
      </div>

      <!-- Footer (projected) -->
      <div class="rc-footer">
        <ng-content />
      </div>
    </cu-card>
  `,
  styles: `
    :host {
      display: block;
    }
    cu-card {
      display: flex;
      flex-direction: column;
      height: 100%;
      padding: 0;
      overflow: hidden;
    }
    .rc-image-wrap {
      width: 100%;
      height: 130px;
      flex-shrink: 0;
      border-radius: var(--cu-radius-lg) var(--cu-radius-lg) 0 0;
      overflow: hidden;
    }
    .rc-image {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }
    .rc-placeholder {
      width: 100%;
      height: 100%;
      background: linear-gradient(135deg, var(--cu-accent), var(--cu-purple));
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .rc-placeholder-letter {
      font-size: 48px;
      font-weight: 700;
      color: #fff;
      line-height: 1;
    }
    .rc-body {
      padding: 16px 20px 12px;
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }
    .rc-name {
      font-size: 16px;
      font-weight: 700;
      color: var(--cu-ink);
      margin: 0;
    }
    .rc-description {
      font-size: 13px;
      color: var(--cu-text-muted);
      margin: 0;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }
    .rc-cost {
      font-size: 14px;
      font-weight: 600;
      color: var(--cu-accent);
      margin: 0;
    }
    .rc-footer {
      padding: 12px 20px 16px;
      display: flex;
      align-items: center;
      gap: 8px;
    }
  `,
})
export class RewardCard {
  readonly name = input.required<string>();
  readonly description = input('');
  readonly cost = input(0);
  readonly imageUrl = input('');

  protected readonly imgFailed = signal(false);
}
