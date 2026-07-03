import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { BootstrapResponse, Collection, InteractableDescription, Prototype, Space, Color, Fragment, AreaDescription } from '../models/editor.models';

@Injectable({ providedIn: 'root' })
export class EditorApiService {
  private readonly http = inject(HttpClient);
  private readonly hostedWorldId = new URLSearchParams(globalThis.location?.search ?? '').get('world');
  private hostedCollection?: Collection;
  private paletteRevision = 0;
  private readonly hostedRevisions = new Map<string, number>();
  private csrfToken = '';

  private get hosted(): boolean { return Boolean(this.hostedWorldId && globalThis.location?.pathname.startsWith('/design')); }

  private async csrf(): Promise<HttpHeaders> {
    if (!this.csrfToken) this.csrfToken = (await firstValueFrom(this.http.get<{ token: string }>('/api/csrf'))).token;
    return new HttpHeaders({ 'X-CSRF-Token': this.csrfToken });
  }

  private revisionKey(kind: string, key: string): string {
    return `${kind}/${key}`;
  }

  private revisionFor(kind: string, key: string): number {
    return this.hostedRevisions.get(this.revisionKey(kind, key)) ?? 0;
  }

  private setRevision(kind: string, key: string, revision: number): void {
    this.hostedRevisions.set(this.revisionKey(kind, key), revision);
  }

  private async saveHosted(kind: string, key: string, content: unknown): Promise<number> {
    const response = await firstValueFrom(this.http.put<{ revision: number }>(`/api/design/worlds/${this.hostedWorldId}/resources/${kind}/${key}`, content, { headers: (await this.csrf()).set('If-Match', String(this.revisionFor(kind, key))) }));
    this.setRevision(kind, key, response.revision);
    return response.revision;
  }

  createCollection(name: string): Promise<void> {
    if (this.hosted) return Promise.reject(new Error('A hosted world contains one collection'));
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
    if (this.hosted) return this.createHostedSpace(input);
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
    if (this.hosted) return this.createHostedArea(input);
    return firstValueFrom(this.http.post<void>('/api/area/create', input));
  }

  getBootstrap(): Promise<BootstrapResponse> {
    if (this.hosted) return this.getHostedBootstrap();
    return firstValueFrom(this.http.get<BootstrapResponse>('/api/bootstrap'));
  }

  saveSpace(collectionName: string, spaceName: string, space: Space): Promise<void> {
    if (this.hosted) {
      this.hostedCollection!.Spaces[spaceName] = space;
      return this.saveHosted('space', spaceName, space).then(() => undefined);
    }
    return firstValueFrom(this.http.put<void>('/api/space', { collectionName, spaceName, space }));
  }

  flattenSpace(collectionName: string, spaceName: string): Promise<{ spaceName: string }> {
    if (this.hosted) return Promise.reject(new Error('Flatten is currently available only in the local workspace'));
    return firstValueFrom(this.http.post<{ spaceName: string }>('/api/space/flatten', { collectionName, spaceName }));
  }

  savePrototypeSet(collectionName: string, setName: string, prototypes: Prototype[]): Promise<void> {
    if (this.hosted) {
      this.hostedCollection!.PrototypeSets[setName] = prototypes;
      return this.saveHosted('prototype-set', setName, prototypes).then(() => undefined);
    }
    return firstValueFrom(this.http.put<void>('/api/prototype-set', { collectionName, setName, prototypes }));
  }

  saveFragmentSet(collectionName: string, setName: string, fragments: Fragment[]): Promise<void> {
    if (this.hosted) {
      this.hostedCollection!.Fragments[setName] = fragments;
      return this.saveHosted('fragment-set', setName, fragments).then(() => undefined);
    }
    return firstValueFrom(this.http.put<void>('/api/fragment-set', { collectionName, setName, fragments }));
  }

  saveInteractableSet(collectionName: string, setName: string, interactables: InteractableDescription[]): Promise<void> {
    if (this.hosted) {
      this.hostedCollection!.InteractableSets[setName] = interactables;
      return this.saveHosted('interactable-set', setName, interactables).then(() => undefined);
    }
    return firstValueFrom(this.http.put<void>('/api/interactable-set', { collectionName, setName, interactables }));
  }

  saveColors(colors: Color[]): Promise<void> {
    if (this.hosted) return this.saveHosted('palette', 'colors', colors).then(revision => { this.paletteRevision = revision; });
    return firstValueFrom(this.http.put<void>('/api/colors', colors));
  }

  async compile(collectionName: string): Promise<void> {
    if (this.hosted) { await firstValueFrom(this.http.post<void>(`/api/design/worlds/${this.hostedWorldId}/releases`, {}, { headers: await this.csrf() })); return; }
    await firstValueFrom(this.http.post<void>('/api/compile', { collectionName }));
  }

  async deploy(collectionName: string): Promise<void> {
    if (this.hosted) { await firstValueFrom(this.http.post<void>(`/api/worlds/${this.hostedWorldId}/launch`, {}, { headers: await this.csrf() })); return; }
    await firstValueFrom(this.http.post<void>('/api/deploy', { collectionName }));
  }

  private async getHostedBootstrap(): Promise<BootstrapResponse> {
    const draft = await firstValueFrom(this.http.get<{ world?: { name?: string }; resources: Array<{ kind: string; key: string; revision: number; content: unknown }> }>(`/api/design/worlds/${this.hostedWorldId}/draft`));
    this.hostedRevisions.clear();
    for (const resource of draft.resources) this.setRevision(resource.kind, resource.key, resource.revision);

    const collectionResource = draft.resources.find(resource => resource.kind === 'collection' && resource.key === 'source');
    if (collectionResource) {
      this.hostedCollection = collectionResource.content as Collection;
    } else {
      const meta = draft.resources.find(resource => resource.kind === 'collection' && resource.key === 'meta');
      const collectionName = ((meta?.content as { name?: string } | undefined)?.name) ?? draft.world?.name ?? 'world';
      const collection: Collection = { Name: collectionName, Spaces: {}, Fragments: {}, PrototypeSets: {}, InteractableSets: {} };
      for (const resource of draft.resources) {
        if (resource.kind === 'space') collection.Spaces[resource.key] = resource.content as Space;
        else if (resource.kind === 'fragment-set') collection.Fragments[resource.key] = resource.content as Fragment[];
        else if (resource.kind === 'prototype-set') collection.PrototypeSets[resource.key] = resource.content as Prototype[];
        else if (resource.kind === 'interactable-set') collection.InteractableSets[resource.key] = resource.content as InteractableDescription[];
      }
      this.hostedCollection = collection;
    }

    const palette = draft.resources.find(resource => resource.kind === 'palette' && resource.key === 'colors');
    this.paletteRevision = palette?.revision ?? 0;
    return { collections: { [this.hostedCollection.Name]: this.hostedCollection }, colors: (palette?.content as Color[] | undefined) ?? [] };
  }

  private async createHostedSpace(input: { collectionName: string; name: string; topology: string; latitude: number; longitude: number; areaWidth: number; areaHeight: number; tileColor: string; tileColor1: string; weather: string; broadcastGroup: string }): Promise<void> {
    const areas: AreaDescription[] = [];
    for (let y = 0; y < input.latitude; y++) for (let x = 0; x < input.longitude; x++) areas.push(this.blankArea(`${input.name}:${y}-${x}`, input.areaHeight, input.areaWidth, input.tileColor, input.tileColor1));
    const space = { CollectionName: input.collectionName, Name: input.name, Topology: input.topology, Latitude: input.latitude, Longitude: input.longitude, AreaHeight: input.areaHeight, AreaWidth: input.areaWidth, Areas: areas };
    this.hostedCollection!.Spaces[input.name] = space;
    await this.saveHosted('space', input.name, space);
  }

  private async createHostedArea(input: { collectionName: string; spaceName: string; name: string; safe: boolean; height: number; width: number; defaultTileColor: string; defaultTileColor1: string }): Promise<void> {
    const area = this.blankArea(input.name, input.height, input.width, input.defaultTileColor, input.defaultTileColor1); area.Safe = input.safe;
    const space = this.hostedCollection!.Spaces[input.spaceName];
    space.Areas.push(area);
    await this.saveHosted('space', input.spaceName, space);
  }

  private blankArea(name: string, height: number, width: number, color: string, color1: string): AreaDescription {
    return { Name: name, Safe: false, Blueprint: { Tiles: Array.from({ length: height }, () => Array.from({ length: width }, () => ({}))), Instructions: [], DefaultTileColor: color, DefaultTileColor1: color1 }, Transports: [] };
  }
}
