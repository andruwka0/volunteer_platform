import { Routes } from '@angular/router';
import { MainLayout } from './widgets/layout/main-layout';
import { authGuard, guestGuard, organizerGuard, adminGuard } from './shared/auth/auth.guards';

export const routes: Routes = [
  {
    path: 'login',
    canActivate: [guestGuard],
    loadComponent: () => import('./pages/login/login').then((m) => m.LoginPage),
  },
  {
    path: 'register',
    canActivate: [guestGuard],
    loadComponent: () => import('./pages/register/register').then((m) => m.RegisterPage),
  },
  {
    path: '',
    component: MainLayout,
    canActivate: [authGuard],
    children: [
      {
        path: 'events',
        loadComponent: () => import('./pages/events/events').then((m) => m.EventsPage),
      },
      {
        path: 'events/:id',
        loadComponent: () =>
          import('./pages/event-detail/event-detail').then((m) => m.EventDetailPage),
      },
      {
        path: 'rewards',
        loadComponent: () => import('./pages/rewards/rewards').then((m) => m.RewardsPage),
      },
      {
        path: 'profile',
        loadComponent: () => import('./pages/profile/profile').then((m) => m.ProfilePage),
      },
      {
        path: 'organizer',
        canActivate: [organizerGuard],
        loadComponent: () => import('./pages/organizer/organizer').then((m) => m.OrganizerPage),
      },
      {
        path: 'admin',
        canActivate: [adminGuard],
        loadComponent: () => import('./pages/admin/admin').then((m) => m.AdminPage),
      },
      { path: '', pathMatch: 'full', redirectTo: 'events' },
    ],
  },
  { path: '**', redirectTo: '' },
];
