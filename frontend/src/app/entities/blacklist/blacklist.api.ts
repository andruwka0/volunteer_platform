import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../shared/api/api.service';
import { ENDPOINTS } from '../../shared/config/api.config';

/** Organizer/Admin blacklist management. */
@Injectable({ providedIn: 'root' })
export class BlacklistApi {
  private readonly api = inject(ApiService);

  ban(userId: number): Observable<{ message: string }> {
    return this.api.post<{ message: string }>(ENDPOINTS.blacklist, { user_id: userId });
  }

  unban(userId: number): Observable<{ message: string }> {
    return this.api.delete<{ message: string }>(ENDPOINTS.unban(userId));
  }
}
