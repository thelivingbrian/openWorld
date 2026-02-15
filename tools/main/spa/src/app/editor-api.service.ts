import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { BootstrapResponse, InteractableDescription, Prototype, Space, Color, Fragment } from './editor.models';

@Injectable({ providedIn: 'root' })
export class EditorApiService {
  private readonly http = inject(HttpClient);

  createCollection(name: string): Promise<void> {
    return firstValueFrom(this.http.post<void>('/api/collection', { name }));
  }

  createSpace(input: {
    collectionName: string;
    name: string;
    topology: string;
    latitude: number;
    longitude: number;
    areaWidth: number;
    areaHeight: number;
    tileColor: string;
    tileColor1: string;
    weather: string;
    broadcastGroup: string;
  }): Promise<void> {
    return firstValueFrom(this.http.post<void>('/api/space/create', input));
  }

  createArea(input: {
    collectionName: string;
    spaceName: string;
    name: string;
    safe: boolean;
    height: number;
    width: number;
    defaultTileColor: string;
    defaultTileColor1: string;
  }): Promise<void> {
    return firstValueFrom(this.http.post<void>('/api/area/create', input));
  }

  getBootstrap(): Promise<BootstrapResponse> {
    return firstValueFrom(this.http.get<BootstrapResponse>('/api/bootstrap'));
  }

  saveSpace(collectionName: string, spaceName: string, space: Space): Promise<void> {
    return firstValueFrom(this.http.put<void>('/api/space', { collectionName, spaceName, space }));
  }

  flattenSpace(collectionName: string, spaceName: string): Promise<{ spaceName: string }> {
    return firstValueFrom(this.http.post<{ spaceName: string }>('/api/space/flatten', { collectionName, spaceName }));
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
