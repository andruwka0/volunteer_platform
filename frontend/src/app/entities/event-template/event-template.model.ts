/** Mirrors EventTemplateDTO. */
export interface EventTemplate {
  id: number;
  title: string;
  description: string;
  location: string;
  cover_image_url: string;
  duration_minutes: number;
  max_participants?: number | null;
  reserve_participants: number;
  skill_points: number;
}

export interface EventTemplateList {
  templates: EventTemplate[];
  count: number;
}

export interface CreateEventTemplatePayload {
  title: string;
  description: string;
  location: string;
  image_url: string;
  duration_minutes: number;
  max_participants?: number | null;
  reserve_participants: number;
  skill_points: number;
}
