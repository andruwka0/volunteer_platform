import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthStore } from '../../shared/auth/auth.store';
import { RegisterPayload } from '../../shared/auth/auth.api';
import { ApiError } from '../../shared/api/api-error';
import { Button, Card, NotificationService } from '../../shared/ui';

@Component({
  selector: 'cu-register-page',
  imports: [ReactiveFormsModule, RouterLink, Button, Card],
  template: `
    <div class="auth-page">
      <cu-card class="auth-card">
        <div class="auth-brand">
          <div class="auth-brand__mark">ЦУ</div>
          <h1 class="auth-brand__title">Центральный · Волонтёрство</h1>
          <p class="auth-brand__subtitle">Создайте новый аккаунт</p>
        </div>

        @if (error()) {
          <div class="cu-alert cu-alert--error" role="alert">{{ error() }}</div>
        }

        <form [formGroup]="form" (ngSubmit)="submit()" novalidate>
          <div class="cu-field">
            <label class="cu-label" for="reg-login">Логин</label>
            <input
              id="reg-login"
              class="cu-input"
              [class.cu-input--invalid]="isInvalid('login')"
              type="text"
              formControlName="login"
              autocomplete="username"
              placeholder="Придумайте логин"
            />
            @if (isInvalid('login')) {
              <span class="cu-field__error">Логин обязателен</span>
            }
          </div>

          <div class="cu-field">
            <label class="cu-label" for="reg-password">Пароль</label>
            <input
              id="reg-password"
              class="cu-input"
              [class.cu-input--invalid]="isInvalid('password')"
              type="password"
              formControlName="password"
              autocomplete="new-password"
              placeholder="Минимум 4 символа"
            />
            @if (isInvalid('password')) {
              @if (form.get('password')?.hasError('required')) {
                <span class="cu-field__error">Пароль обязателен</span>
              } @else if (form.get('password')?.hasError('minlength')) {
                <span class="cu-field__error">Пароль должен быть не менее 4 символов</span>
              }
            }
          </div>

          <div class="cu-field">
            <label class="cu-label" for="reg-last-name">Фамилия</label>
            <input
              id="reg-last-name"
              class="cu-input"
              [class.cu-input--invalid]="isInvalid('last_name')"
              type="text"
              formControlName="last_name"
              autocomplete="family-name"
              placeholder="Введите фамилию"
            />
            @if (isInvalid('last_name')) {
              <span class="cu-field__error">Фамилия обязательна</span>
            }
          </div>

          <div class="cu-field">
            <label class="cu-label" for="reg-first-name">Имя</label>
            <input
              id="reg-first-name"
              class="cu-input"
              [class.cu-input--invalid]="isInvalid('first_name')"
              type="text"
              formControlName="first_name"
              autocomplete="given-name"
              placeholder="Введите имя"
            />
            @if (isInvalid('first_name')) {
              <span class="cu-field__error">Имя обязательно</span>
            }
          </div>

          <div class="cu-field">
            <label class="cu-label" for="reg-middle-name">Отчество <span class="optional">(необязательно)</span></label>
            <input
              id="reg-middle-name"
              class="cu-input"
              type="text"
              formControlName="middle_name"
              autocomplete="additional-name"
              placeholder="Введите отчество"
            />
          </div>

          <div class="cu-field">
            <label class="cu-label" for="reg-telegram">Telegram <span class="optional">(необязательно)</span></label>
            <input
              id="reg-telegram"
              class="cu-input"
              type="text"
              formControlName="telegram"
              placeholder="@username"
            />
          </div>

          <cu-button
            type="submit"
            [variant]="'primary'"
            [loading]="submitting()"
            [full]="true"
          >
            Зарегистрироваться
          </cu-button>
        </form>

        <p class="auth-link">
          Уже есть аккаунт?
          <a routerLink="/login">Войти</a>
        </p>
      </cu-card>
    </div>
  `,
  styles: `
    .auth-page {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--cu-surface-2);
      padding: 24px;
    }

    .auth-card {
      width: 100%;
      max-width: 440px;
    }

    .auth-brand {
      text-align: center;
      margin-bottom: 28px;
    }

    .auth-brand__mark {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 56px;
      height: 56px;
      border-radius: 16px;
      background: var(--cu-ink);
      color: var(--cu-text-on-dark);
      font-size: 20px;
      font-weight: 700;
      margin-bottom: 12px;
    }

    .auth-brand__title {
      font-size: 20px;
      font-weight: 700;
      color: var(--cu-text);
      margin: 0 0 4px;
    }

    .auth-brand__subtitle {
      font-size: 14px;
      color: var(--cu-text-2);
      margin: 0;
    }

    .cu-alert--error {
      border-radius: 8px;
      padding: 10px 14px;
      font-size: 14px;
      margin-bottom: 16px;
      background: var(--cu-danger-soft);
      color: var(--cu-danger);
      border: 1px solid var(--cu-danger-border, var(--cu-danger));
    }

    form {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }

    .optional {
      font-size: 12px;
      color: var(--cu-text-3);
      font-weight: 400;
    }

    .auth-link {
      margin-top: 20px;
      text-align: center;
      font-size: 14px;
      color: var(--cu-text-2);
    }

    .auth-link a {
      color: var(--cu-accent);
      text-decoration: none;
      font-weight: 500;
    }

    .auth-link a:hover {
      text-decoration: underline;
    }
  `,
})
export class RegisterPage {
  private readonly fb = inject(FormBuilder);
  private readonly authStore = inject(AuthStore);
  private readonly router = inject(Router);
  private readonly notifications = inject(NotificationService);

  protected readonly submitting = signal(false);
  protected readonly error = signal<string | null>(null);

  protected readonly form = this.fb.group({
    login: ['', Validators.required],
    password: ['', [Validators.required, Validators.minLength(4)]],
    first_name: ['', Validators.required],
    last_name: ['', Validators.required],
    middle_name: [''],
    telegram: [''],
  });

  protected isInvalid(field: keyof typeof this.form.controls): boolean {
    const ctrl = this.form.get(field as string);
    return !!(ctrl && ctrl.invalid && ctrl.touched);
  }

  protected submit(): void {
    if (this.submitting()) return;
    this.form.markAllAsTouched();
    if (this.form.invalid) return;

    const raw = this.form.getRawValue();
    const payload: RegisterPayload = {
      login: raw.login ?? '',
      password: raw.password ?? '',
      first_name: raw.first_name ?? '',
      last_name: raw.last_name ?? '',
      middle_name: raw.middle_name ?? '',
      telegram: raw.telegram ?? '',
    };

    this.submitting.set(true);
    this.error.set(null);

    this.authStore.register(payload).subscribe({
      next: () => {
        this.submitting.set(false);
        void this.router.navigateByUrl('/events');
      },
      error: (e: ApiError) => {
        this.submitting.set(false);
        this.error.set(e.message);
        this.notifications.error(e.message);
      },
    });
  }
}
