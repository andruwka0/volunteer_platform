import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../shared/api/api.service';
import { ENDPOINTS } from '../../shared/config/api.config';
import { UserEventHistoryList } from '../participant/participant.model';

@Injectable({ providedIn: 'root' })
export class MeApi {
  private readonly api = inject(ApiService);

  /** Current user's event participation history. */
  events(): Observable<UserEventHistoryList> {
    return this.api.get<UserEventHistoryList>(ENDPOINTS.myEvents);
  }
}
