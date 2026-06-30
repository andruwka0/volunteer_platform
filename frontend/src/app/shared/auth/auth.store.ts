import { Injectable, computed, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, catchError, switchMap, tap, throwError } from 'rxjs';
import { TokenStorage } from './token-storage';
import { AuthApi, RegisterPayload } from './auth.api';
import { UserApi } from '../../entities/user/user.api';
import { User, UserRole } from '../../entities/user/user.model';

/** Holds authentication state (token + current user) as signals. */
@Injectable({ providedIn: 'root' })
export class AuthStore {
  private readonly tokenStorage = inject(TokenStorage);
  private readonly authApi = inject(AuthApi);
  private readonly userApi = inject(UserApi);
  private readonly router = inject(Router);

  private readonly _user = signal<User | null>(null);
  private readonly _loading = signal(false);

  readonly user = this._user.asReadonly();
  readonly loading = this._loading.asReadonly();
  readonly isAuthenticated = computed(() => !!this.tokenStorage.token());
  readonly role = computed<UserRole | null>(() => this._user()?.role ?? null);
  readonly isAdmin = computed(() => this.role() === 'Admin');
  readonly isOrganizer = computed(() => {
    const r = this.role();
    return r === 'Organizer' || r === 'Admin';
  });
  readonly skillPoints = computed(() => this._user()?.skill_points ?? 0);

  login(login: string, password: string): Observable<User> {
    return this.authApi.login(login, password).pipe(
      tap((res) => this.tokenStorage.set(res.token)),
      switchMap(() => this.refresh()),
    );
  }

  register(payload: RegisterPayload): Observable<User> {
    return this.authApi.register(payload).pipe(
      tap((res) => this.tokenStorage.set(res.token)),
      switchMap(() => this.refresh()),
    );
  }

  /** Re-fetch the current user from /auth/me. */
  refresh(): Observable<User> {
    this._loading.set(true);
    return this.userApi.me().pipe(
      tap((user) => {
        this._user.set(user);
        this._loading.set(false);
      }),
      catchError((err) => {
        this._loading.set(false);
        return throwError(() => err);
      }),
    );
  }

  setUser(user: User): void {
    this._user.set(user);
  }

  logout(): void {
    this.tokenStorage.clear();
    this._user.set(null);
    void this.router.navigateByUrl('/login');
  }
}
