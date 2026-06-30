import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../shared/api/api.service';
import { ENDPOINTS } from '../../shared/config/api.config';
import { CreateEventPayload, EventList, EventModel } from './event.model';
import { ParticipantList } from '../participant/participant.model';

@Injectable({ providedIn: 'root' })
export class EventApi {
  private readonly api = inject(ApiService);

  list(): Observable<EventList> {
    return this.api.get<EventList>(ENDPOINTS.events);
  }

  byId(id: number): Observable<EventModel> {
    return this.api.get<EventModel>(ENDPOINTS.event(id));
  }

  participants(id: number): Observable<ParticipantList> {
    return this.api.get<ParticipantList>(ENDPOINTS.eventParticipants(id));
  }

  create(payload: CreateEventPayload): Observable<EventModel> {
    return this.api.post<EventModel>(ENDPOINTS.events, payload);
  }

  register(id: number): Observable<{ message: string }> {
    return this.api.post<{ message: string }>(ENDPOINTS.eventRegister(id));
  }

  cancelRegistration(id: number): Observable<{ message: string }> {
    return this.api.delete<{ message: string }>(ENDPOINTS.eventRegister(id));
  }

  /** Organizer: confirm a participant attended. */
  confirmAttendance(eventId: number, userId: number): Observable<{ message: string }> {
    return this.api.post<{ message: string }>(ENDPOINTS.confirmAttendance(eventId, userId));
  }

  /** Admin: approve a FINISHED event and award skill points. */
  approve(id: number): Observable<{ message: string }> {
    return this.api.post<{ message: string }>(ENDPOINTS.approveEvent(id));
  }
}
