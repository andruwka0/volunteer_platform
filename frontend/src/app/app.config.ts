import {
  ApplicationConfig,
  provideBrowserGlobalErrorListeners,
  provideZonelessChangeDetection,
  provideAppInitializer,
  inject,
} from '@angular/core';
import { provideRouter, withComponentInputBinding } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { catchError, firstValueFrom, of } from 'rxjs';

import { routes } from './app.routes';
import { authInterceptor } from './shared/auth/auth.interceptor';
import { TokenStorage } from './shared/auth/token-storage';
import { AuthStore } from './shared/auth/auth.store';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideZonelessChangeDetection(),
    provideHttpClient(withInterceptors([authInterceptor])),
    provideRouter(routes, withComponentInputBinding()),
    // Eagerly load the current user when a token is present so role guards work.
    provideAppInitializer(() => {
      const tokenStorage = inject(TokenStorage);
      const authStore = inject(AuthStore);
      if (!tokenStorage.token()) return;
      return firstValueFrom(authStore.refresh().pipe(catchError(() => of(null))));
    }),
  ],
};
