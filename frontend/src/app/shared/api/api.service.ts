import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpErrorResponse, HttpParams } from '@angular/common/http';
import { Observable, catchError, map, throwError } from 'rxjs';
import { ApiError } from './api-error';

/** Standard backend response envelope: { success, data, error }. */
export interface ApiEnvelope<T> {
  success: boolean;
  data?: T;
  error?: { code: string; message: string };
}

type QueryParams = Record<string, string | number | boolean | undefined | null>;

/**
 * Thin wrapper over HttpClient that unwraps the `{ success, data, error }`
 * envelope and normalizes every failure into an {@link ApiError}.
 */
@Injectable({ providedIn: 'root' })
export class ApiService {
  private readonly http = inject(HttpClient);

  get<T>(url: string, params?: QueryParams): Observable<T> {
    return this.unwrap(this.http.get<ApiEnvelope<T>>(url, { params: this.toParams(params) }));
  }

  post<T>(url: string, body?: unknown): Observable<T> {
    return this.unwrap(this.http.post<ApiEnvelope<T>>(url, body ?? {}));
  }

  delete<T>(url: string, body?: unknown): Observable<T> {
    return this.unwrap(
      this.http.request<ApiEnvelope<T>>('DELETE', url, { body: body ?? {} }),
    );
  }

  private unwrap<T>(source: Observable<ApiEnvelope<T>>): Observable<T> {
    return source.pipe(
      map((res) => {
        if (res && !res.success) {
          throw new ApiError(
            res.error?.code ?? 'UNKNOWN',
            res.error?.message ?? 'Неизвестная ошибка',
            200,
          );
        }
        return (res?.data ?? (undefined as unknown)) as T;
      }),
      catchError((err: unknown) => throwError(() => this.toApiError(err))),
    );
  }

  private toApiError(err: unknown): ApiError {
    if (err instanceof ApiError) {
      return err;
    }
    if (err instanceof HttpErrorResponse) {
      const body = err.error;
      // JSON envelope with structured error
      if (body && typeof body === 'object' && body.error?.message) {
        return new ApiError(body.error.code ?? 'ERROR', body.error.message, err.status);
      }
      // Plain-text body (e.g. middleware "forbidden" / "unauthorized")
      const text = typeof body === 'string' && body.trim() ? body.trim() : null;
      const message =
        text ??
        (err.status === 401
          ? 'Требуется авторизация'
          : err.status === 403
            ? 'Недостаточно прав'
            : err.status === 0
              ? 'Сервер недоступен'
              : `Ошибка ${err.status}`);
      return new ApiError(`HTTP_${err.status}`, message, err.status);
    }
    return new ApiError('UNKNOWN', 'Неизвестная ошибка', 0);
  }

  private toParams(params?: QueryParams): HttpParams {
    let httpParams = new HttpParams();
    if (params) {
      for (const [key, value] of Object.entries(params)) {
        if (value !== undefined && value !== null && value !== '') {
          httpParams = httpParams.set(key, String(value));
        }
      }
    }
    return httpParams;
  }
}
