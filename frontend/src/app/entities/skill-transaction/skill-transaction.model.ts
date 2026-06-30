export type TransactionType = 'manual' | 'event' | 'reward';

/**
 * Mirrors domain.SkillPointTransaction. NOTE: this endpoint serializes the raw
 * Go struct WITHOUT json tags, so the keys are PascalCase (ID, Points, ...).
 */
export interface SkillPointTransaction {
  ID: number;
  UserID: number;
  Points: number;
  Type: TransactionType | string;
  Reason: string;
  EventID?: number | null;
  CreatedAt: string;
}

export interface SkillPointHistory {
  user_id: number;
  transactions: SkillPointTransaction[];
  count: number;
}

export const TRANSACTION_TYPE_LABELS: Record<string, string> = {
  manual: 'Начисление вручную',
  event: 'За мероприятие',
  reward: 'Покупка награды',
};
