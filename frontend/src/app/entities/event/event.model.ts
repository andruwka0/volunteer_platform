export type EventStatus =
  | 'EVENT-RECRUITING'
  | 'EVENT-ACTIVE'
  | 'EVENT-FINISHED'
  | 'EVENT-CLOSED'
  | 'EVENT-CANCELLED';

/** Mirrors EventDTO. */
export interface EventModel {
  id: number;
  title: string;
  description: string;
  location: string;
  cover_image_url: string;
  status: EventStatus;
  start_date: string;
  end_date: string;
  registration_deadline?: string | null;
  max_participants?: number | null;
  reserve_participants: number;
  skill_points: number;
  created_by_id: number;
  participants_count: number;
  reserve_count: number;
}

export interface EventList {
  events: EventModel[];
  count: number;
}

/** Payload for POST /events. Dates must be RFC3339. */
export interface CreateEventPayload {
  title: string;
  description: string;
  location: string;
  image: string;
  start_date: string;
  end_date: string;
  registration_deadline?: string | null;
  max_participants?: number | null;
  reserve_participants: number;
  skill_points: number;
  template_id?: number | null;
}

export const EVENT_STATUS_LABELS: Record<EventStatus, string> = {
  'EVENT-RECRUITING': 'Набор открыт',
  'EVENT-ACTIVE': 'Идёт',
  'EVENT-FINISHED': 'Завершён',
  'EVENT-CLOSED': 'Закрыт',
  'EVENT-CANCELLED': 'Отменён',
};

/** Maps status -> ui-kit badge tone. */
export const EVENT_STATUS_TONE: Record<EventStatus, 'success' | 'info' | 'warning' | 'neutral' | 'danger'> = {
  'EVENT-RECRUITING': 'success',
  'EVENT-ACTIVE': 'info',
  'EVENT-FINISHED': 'warning',
  'EVENT-CLOSED': 'neutral',
  'EVENT-CANCELLED': 'danger',
};
