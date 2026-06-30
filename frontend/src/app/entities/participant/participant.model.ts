/** Mirrors ParticipantDTO. */
export interface Participant {
  user_id: number;
  first_name: string;
  last_name: string;
  middle_name: string;
  telegram: string;
  attendance_confirmed: boolean;
}

export interface ParticipantList {
  event_id: number;
  participants: Participant[];
  count: number;
}

/** Mirrors UserEventHistoryDTO (an event in a user's participation history). */
export interface UserEventHistory {
  event_id: number;
  title: string;
  status: string;
  joined_at: string;
  attendance_confirmed: boolean;
  skill_points: number;
}

export interface UserEventHistoryList {
  user_id: number;
  events: UserEventHistory[];
  count: number;
}
