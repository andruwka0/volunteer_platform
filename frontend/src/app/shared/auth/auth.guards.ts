import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { TokenStorage } from './token-storage';
import { AuthStore } from './auth.store';

/** Requires a valid session (token present). */
export const authGuard: CanActivateFn = () => {
  const token = inject(TokenStorage).token();
  const router = inject(Router);
  return token ? true : router.parseUrl('/login');
};

/** Redirects authenticated users away from login/register. */
export const guestGuard: CanActivateFn = () => {
  const token = inject(TokenStorage).token();
  const router = inject(Router);
  return token ? router.parseUrl('/events') : true;
};

/** Requires Organizer or Admin role. */
export const organizerGuard: CanActivateFn = () => {
  const store = inject(AuthStore);
  const router = inject(Router);
  return store.isOrganizer() ? true : router.parseUrl('/events');
};

/** Requires Admin role. */
export const adminGuard: CanActivateFn = () => {
  const store = inject(AuthStore);
  const router = inject(Router);
  return store.isAdmin() ? true : router.parseUrl('/events');
};
