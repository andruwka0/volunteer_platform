import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../shared/api/api.service';
import { ENDPOINTS } from '../../shared/config/api.config';
import { CreateRewardPayload, Reward, UserRewardList } from './reward.model';

@Injectable({ providedIn: 'root' })
export class RewardApi {
  private readonly api = inject(ApiService);

  /** Global reward catalog. GET /admin/rewards returns a bare array. */
  list(): Observable<Reward[]> {
    return this.api.get<Reward[]>(ENDPOINTS.rewards);
  }

  /** Admin: create a reward. */
  create(payload: CreateRewardPayload): Observable<Reward> {
    return this.api.post<Reward>(ENDPOINTS.rewards, payload);
  }

  /** Current user's rewards (purchased / available). */
  myRewards(): Observable<UserRewardList> {
    return this.api.get<UserRewardList>(ENDPOINTS.myRewards);
  }

  /** Current user: claim (reserve) a reward by spending skill points. */
  claim(rewardId: number): Observable<{ message: string }> {
    return this.api.post<{ message: string }>(ENDPOINTS.claimReward(rewardId));
  }

  /** Admin: confirm a user picked up their merch. */
  confirmPickup(userId: number, rewardId: number): Observable<{ message: string }> {
    return this.api.post<{ message: string }>(ENDPOINTS.confirmPickup(userId, rewardId));
  }
}
