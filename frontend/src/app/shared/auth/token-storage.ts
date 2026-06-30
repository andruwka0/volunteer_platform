import { Injectable, signal } from '@angular/core';

/** Persists the JWT in localStorage and exposes it as a signal. */
@Injectable({ providedIn: 'root' })
export class TokenStorage {
  private readonly KEY = 'cu_token';
  private readonly _token = signal<string | null>(this.read());
  readonly token = this._token.asReadonly();

  set(token: string): void {
    try {
      localStorage.setItem(this.KEY, token);
    } catch {
      /* ignore storage errors */
    }
    this._token.set(token);
  }

  clear(): void {
    try {
      localStorage.removeItem(this.KEY);
    } catch {
      /* ignore */
    }
    this._token.set(null);
  }

  private read(): string | null {
    try {
      return localStorage.getItem(this.KEY);
    } catch {
      return null;
    }
  }
}
