import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { BootstrapResponse, InteractableDescription, Prototype, Space, Color, Fragment } from './editor.models';

@Injectable({ providedIn: 'root' })
export class EditorApiService {
  private readonly http = inject(HttpClient);

  getBootstrap(): Promise<BootstrapResponse> {
    return firstValueFrom(this.http.get<BootstrapResponse>('/api/bootstrap'));
  }

  saveSpace(collectionName: string, spaceName: string, space: Space): Promise<void> {
    return firstValueFrom(this.http.put<void>('/api/space', { collectionName, spaceName, space }));
  }

  savePrototypeSet(collectionName: string, setName: string, prototypes: Prototype[]): Promise<void> {
    return firstValueFrom(this.http.put<void>('/api/prototype-set', { collectionName, setName, prototypes }));
  }

  saveFragmentSet(collectionName: string, setName: string, fragments: Fragment[]): Promise<void> {
    return firstValueFrom(this.http.put<void>('/api/fragment-set', { collectionName, setName, fragments }));
  }

  saveInteractableSet(collectionName: string, setName: string, interactables: InteractableDescription[]): Promise<void> {
    return firstValueFrom(this.http.put<void>('/api/interactable-set', { collectionName, setName, interactables }));
  }

  saveColors(colors: Color[]): Promise<void> {
    return firstValueFrom(this.http.put<void>('/api/colors', colors));
  }

  compile(collectionName: string): Promise<void> {
    return firstValueFrom(this.http.post<void>('/api/compile', { collectionName }));
  }

  deploy(collectionName: string): Promise<void> {
    return firstValueFrom(this.http.post<void>('/api/deploy', { collectionName }));
  }
}
