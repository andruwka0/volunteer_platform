import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../api/api.service';
import { ENDPOINTS } from '../config/api.config';

export interface RegisterPayload {
  login: string;
  password: string;
  first_name: string;
  last_name: string;
  middle_name: string;
  telegram: string;
}

interface TokenResponse {
  token: string;
}

@Injectable({ providedIn: 'root' })
export class AuthApi {
  private readonly api = inject(ApiService);

  login(login: string, password: string): Observable<TokenResponse> {
    return this.api.post<TokenResponse>(ENDPOINTS.login, { login, password });
  }

  register(payload: RegisterPayload): Observable<TokenResponse> {
    return this.api.post<TokenResponse>(ENDPOINTS.register, payload);
  }
}
