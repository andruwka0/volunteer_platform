/**
 * Central place for API base + endpoint builders.
 * In dev, requests are relative and proxied to localhost:8080 (proxy.conf.json).
 */
export const API_BASE = '/api';

export const ENDPOINTS = {
  // auth
  register: `${API_BASE}/auth/register`,
  login: `${API_BASE}/auth/login`,
  me: `${API_BASE}/auth/me`,

  // current user
  myEvents: `${API_BASE}/users/me/events`,
  myRewards: `${API_BASE}/users/me/rewards`,
  claimReward: (rewardId: number) => `${API_BASE}/users/me/rewards/${rewardId}/claim`,
  skillPointHistory: (userId: number) => `${API_BASE}/users/${userId}/skill-points/history`,

  // events
  events: `${API_BASE}/events`,
  event: (id: number) => `${API_BASE}/events/${id}`,
  eventParticipants: (id: number) => `${API_BASE}/events/${id}/participants`,
  eventRegister: (id: number) => `${API_BASE}/events/${id}/register`,

  // organizer
  confirmAttendance: (eventId: number, userId: number) =>
    `${API_BASE}/organizer/events/${eventId}/users/${userId}/confirm`,
  blacklist: `${API_BASE}/organizer/blacklist`,
  unban: (userId: number) => `${API_BASE}/organizer/blacklist/${userId}`,
  eventTemplates: `${API_BASE}/organizer/event-templates`,

  // admin
  adminUsers: `${API_BASE}/admin/users`,
  promoteUser: (id: number) => `${API_BASE}/admin/users/${id}/promote`,
  awardUser: (id: number) => `${API_BASE}/admin/users/${id}/award`,
  adminUserRewards: (id: number) => `${API_BASE}/admin/users/${id}/rewards`,
  approveEvent: (id: number) => `${API_BASE}/admin/events/${id}/approve`,
  rewards: `${API_BASE}/admin/rewards`,
  confirmPickup: (userId: number, rewardId: number) =>
    `${API_BASE}/admin/users/${userId}/rewards/${rewardId}/pickup`,

  // assets
  images: `${API_BASE}/assets/images`,
} as const;
