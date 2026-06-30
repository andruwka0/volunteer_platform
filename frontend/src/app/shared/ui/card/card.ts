import { Component, input } from '@angular/core';

/** White rounded surface used throughout the app. */
@Component({
  selector: 'cu-card',
  template: `<ng-content />`,
  host: {
    '[class.cu-card--interactive]': 'interactive()',
    '[class.cu-card--flat]': 'flat()',
    class: 'cu-card',
  },
  styles: `
    :host {
      display: block;
      background: var(--cu-surface);
      border: 1px solid var(--cu-border);
      border-radius: var(--cu-radius-lg);
      box-shadow: var(--cu-shadow-sm);
      padding: 20px;
    }
    :host(.cu-card--flat) {
      box-shadow: none;
    }
    :host(.cu-card--interactive) {
      cursor: pointer;
      transition:
        box-shadow 0.15s ease,
        transform 0.1s ease,
        border-color 0.15s ease;
    }
    :host(.cu-card--interactive:hover) {
      box-shadow: var(--cu-shadow);
      border-color: var(--cu-border-strong);
      transform: translateY(-2px);
    }
  `,
})
export class Card {
  readonly interactive = input(false);
  readonly flat = input(false);
}
