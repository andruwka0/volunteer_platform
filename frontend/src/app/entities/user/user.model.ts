export type UserRole = 'Admin' | 'Volunteer' | 'Organizer';

/** Mirrors UserDTO (snake_case wire format). */
export interface User {
  id: number;
  login: string;
  first_name: string;
  last_name: string;
  middle_name: string;
  telegram: string;
  role: UserRole;
  skill_points: number;
}

export interface UserSearchResult {
  users: User[];
  count: number;
}

export const ROLE_LABELS: Record<UserRole, string> = {
  Admin: 'Администратор',
  Organizer: 'Организатор',
  Volunteer: 'Волонтёр',
};

export function fullName(u: Pick<User, 'first_name' | 'last_name' | 'middle_name'>): string {
  return [u.last_name, u.first_name, u.middle_name].filter(Boolean).join(' ').trim();
}

export function initials(u: Pick<User, 'first_name' | 'last_name'>): string {
  return `${(u.first_name?.[0] ?? '').toUpperCase()}${(u.last_name?.[0] ?? '').toUpperCase()}`;
}
