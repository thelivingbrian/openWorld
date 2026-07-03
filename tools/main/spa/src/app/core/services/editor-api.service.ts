import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { BootstrapResponse, Collection, InteractableDescription, Prototype, Space, Color, Fragment, AreaDescription } from '../models/editor.models';

@Injectable({ providedIn: 'root' })
export class EditorApiService {
  private readonly http = inject(HttpClient);
  private readonly hostedWorldId = new URLSearchParams(globalThis.location?.search ?? '').get('world');
  private hostedCollection?: Collection;
  private collectionRevision = 0;
  private paletteRevision = 0;
  private csrfToken = '';

  private get hosted(): boolean { return Boolean(this.hostedWorldId && globalThis.location?.pathname.startsWith('/design')); }

  private async csrf(): Promise<HttpHeaders> {
    if (!this.csrfToken) this.csrfToken = (await firstValueFrom(this.http.get<{ token: string }>('/api/csrf'))).token;
    return new HttpHeaders({ 'X-CSRF-Token': this.csrfToken });
  }

  private async saveHosted(kind: string, key: string, content: unknown, revision: number): Promise<number> {
    const response = await firstValueFrom(this.http.put<{ revision: number }>(`/api/design/worlds/${this.hostedWorldId}/resources/${kind}/${key}`, content, { headers: (await this.csrf()).set('If-Match', String(revision)) }));
    return response.revision;
  }

  private async saveHostedCollection(): Promise<void> {
    if (!this.hostedCollection) throw new Error('Hosted collection is not loaded');
    this.collectionRevision = await this.saveHosted('collection', 'source', this.hostedCollection, this.collectionRevision);
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
    if (this.hosted) { this.hostedCollection!.Spaces[spaceName] = space; return this.saveHostedCollection(); }
    return firstValueFrom(this.http.put<void>('/api/space', { collectionName, spaceName, space }));
  }

  flattenSpace(collectionName: string, spaceName: string): Promise<{ spaceName: string }> {
    if (this.hosted) return Promise.reject(new Error('Flatten is currently available only in the local workspace'));
    return firstValueFrom(this.http.post<{ spaceName: string }>('/api/space/flatten', { collectionName, spaceName }));
  }

  savePrototypeSet(collectionName: string, setName: string, prototypes: Prototype[]): Promise<void> {
    if (this.hosted) { this.hostedCollection!.PrototypeSets[setName] = prototypes; return this.saveHostedCollection(); }
    return firstValueFrom(this.http.put<void>('/api/prototype-set', { collectionName, setName, prototypes }));
  }

  saveFragmentSet(collectionName: string, setName: string, fragments: Fragment[]): Promise<void> {
    if (this.hosted) { this.hostedCollection!.Fragments[setName] = fragments; return this.saveHostedCollection(); }
    return firstValueFrom(this.http.put<void>('/api/fragment-set', { collectionName, setName, fragments }));
  }

  saveInteractableSet(collectionName: string, setName: string, interactables: InteractableDescription[]): Promise<void> {
    if (this.hosted) { this.hostedCollection!.InteractableSets[setName] = interactables; return this.saveHostedCollection(); }
    return firstValueFrom(this.http.put<void>('/api/interactable-set', { collectionName, setName, interactables }));
  }

  saveColors(colors: Color[]): Promise<void> {
    if (this.hosted) return this.saveHosted('palette', 'colors', colors, this.paletteRevision).then(revision => { this.paletteRevision = revision; });
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
    const draft = await firstValueFrom(this.http.get<{ resources: Array<{ kind: string; key: string; revision: number; content: unknown }> }>(`/api/design/worlds/${this.hostedWorldId}/draft`));
    const collectionResource = draft.resources.find(resource => resource.kind === 'collection' && resource.key === 'source');
    if (!collectionResource) throw new Error('Hosted world has no collection/source resource');
    this.hostedCollection = collectionResource.content as Collection;
    this.collectionRevision = collectionResource.revision;
    const palette = draft.resources.find(resource => resource.kind === 'palette' && resource.key === 'colors');
    this.paletteRevision = palette?.revision ?? 0;
    return { collections: { [this.hostedCollection.Name]: this.hostedCollection }, colors: (palette?.content as Color[] | undefined) ?? [] };
  }

  private async createHostedSpace(input: { collectionName: string; name: string; topology: string; latitude: number; longitude: number; areaWidth: number; areaHeight: number; tileColor: string; tileColor1: string; weather: string; broadcastGroup: string }): Promise<void> {
    const areas: AreaDescription[] = [];
    for (let y = 0; y < input.latitude; y++) for (let x = 0; x < input.longitude; x++) areas.push(this.blankArea(`${input.name}:${y}-${x}`, input.areaHeight, input.areaWidth, input.tileColor, input.tileColor1));
    this.hostedCollection!.Spaces[input.name] = { CollectionName: input.collectionName, Name: input.name, Topology: input.topology, Latitude: input.latitude, Longitude: input.longitude, AreaHeight: input.areaHeight, AreaWidth: input.areaWidth, Areas: areas };
    await this.saveHostedCollection();
  }

  private async createHostedArea(input: { collectionName: string; spaceName: string; name: string; safe: boolean; height: number; width: number; defaultTileColor: string; defaultTileColor1: string }): Promise<void> {
    const area = this.blankArea(input.name, input.height, input.width, input.defaultTileColor, input.defaultTileColor1); area.Safe = input.safe;
    this.hostedCollection!.Spaces[input.spaceName].Areas.push(area); await this.saveHostedCollection();
  }

  private blankArea(name: string, height: number, width: number, color: string, color1: string): AreaDescription {
    return { Name: name, Safe: false, Blueprint: { Tiles: Array.from({ length: height }, () => Array.from({ length: width }, () => ({}))), Instructions: [], DefaultTileColor: color, DefaultTileColor1: color1 }, Transports: [] };
  }
}
