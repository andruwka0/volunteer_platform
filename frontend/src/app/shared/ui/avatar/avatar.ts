import { Component, input } from '@angular/core';

/** Round avatar with initials fallback. */
@Component({
  selector: 'cu-avatar',
  template: `
    <span class="cu-avatar" [style.width.px]="size()" [style.height.px]="size()">
      {{ initials() }}
    </span>
  `,
  styles: `
    :host {
      display: flex;
    }
    .cu-avatar {
      display: inline-flex;
      margin: auto 0;
      align-items: center;
      justify-content: center;
      border-radius: 50%;
      background: linear-gradient(135deg, var(--cu-accent), var(--cu-purple));
      color: #fff;
      font-weight: 700;
      font-size: 0.42em;
      text-transform: uppercase;
      user-select: none;
      flex-shrink: 0;
    }
  `,
  host: {
    '[style.font-size.px]': 'size()',
  },
})
export class Avatar {
  readonly initials = input('?');
  readonly size = input(40);
}
