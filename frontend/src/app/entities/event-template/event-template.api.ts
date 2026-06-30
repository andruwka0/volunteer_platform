import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../shared/api/api.service';
import { ENDPOINTS } from '../../shared/config/api.config';
import {
  CreateEventTemplatePayload,
  EventTemplate,
  EventTemplateList,
} from './event-template.model';

@Injectable({ providedIn: 'root' })
export class EventTemplateApi {
  private readonly api = inject(ApiService);

  list(): Observable<EventTemplateList> {
    return this.api.get<EventTemplateList>(ENDPOINTS.eventTemplates);
  }

  create(payload: CreateEventTemplatePayload): Observable<EventTemplate> {
    return this.api.post<EventTemplate>(ENDPOINTS.eventTemplates, payload);
  }
}
