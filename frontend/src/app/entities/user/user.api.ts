import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../shared/api/api.service';
import { ENDPOINTS } from '../../shared/config/api.config';
import { User, UserSearchResult } from './user.model';
import { UserRewardList } from '../reward/reward.model';
import { SkillPointHistory } from '../skill-transaction/skill-transaction.model';

@Injectable({ providedIn: 'root' })
export class UserApi {
  private readonly api = inject(ApiService);

  me(): Observable<User> {
    return this.api.get<User>(ENDPOINTS.me);
  }

  /** Admin: search users. first_name & last_name are required by the backend. */
  search(firstName: string, lastName: string, middleName?: string): Observable<UserSearchResult> {
    return this.api.get<UserSearchResult>(ENDPOINTS.adminUsers, {
      first_name: firstName,
      last_name: lastName,
      middle_name: middleName,
    });
  }

  promote(userId: number, role: User['role']): Observable<{ message: string }> {
    return this.api.post<{ message: string }>(ENDPOINTS.promoteUser(userId), { role });
  }

  award(userId: number, points: number, reason: string): Observable<{ message: string }> {
    return this.api.post<{ message: string }>(ENDPOINTS.awardUser(userId), { points, reason });
  }

  /** Admin: a specific user's claimed rewards (for merch pickup). */
  rewardsByAdmin(userId: number): Observable<UserRewardList> {
    return this.api.get<UserRewardList>(ENDPOINTS.adminUserRewards(userId));
  }

  skillPointHistory(userId: number): Observable<SkillPointHistory> {
    return this.api.get<SkillPointHistory>(ENDPOINTS.skillPointHistory(userId));
  }
}
