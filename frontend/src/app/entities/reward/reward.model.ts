/** Mirrors RewardDTO (global reward catalog item). */
export interface Reward {
  id: number;
  name: string;
  description: string;
  cost: number;
  image_url: string;
}

/** Mirrors UserRewardDTO (a reward in the context of a specific user). */
export interface UserReward {
  reward_id: number;
  name: string;
  description: string;
  cost: number;
  image_url: string;
  available: boolean;
  claimed: boolean;
  picked_up: boolean;
}

export interface UserRewardList {
  user_id: number;
  rewards: UserReward[];
}

export interface CreateRewardPayload {
  name: string;
  description: string;
  image_url: string;
  cost: number;
}
