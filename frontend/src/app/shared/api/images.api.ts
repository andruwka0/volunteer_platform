import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from './api.service';
import { ENDPOINTS } from '../config/api.config';

export interface ImageList {
  images: string[];
  count: number;
}

@Injectable({ providedIn: 'root' })
export class ImagesApi {
  private readonly api = inject(ApiService);

  list(): Observable<ImageList> {
    return this.api.get<ImageList>(ENDPOINTS.images);
  }
}
