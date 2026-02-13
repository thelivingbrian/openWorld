import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { EditorApiService } from './editor-api.service';
import {
  AreaDescription,
  BootstrapResponse,
  Collection,
  Fragment,
  GridSelection,
  InteractableDescription,
  Material,
  Prototype,
  Space,
} from './editor.models';
import { applyGridTool, generateMaterials, Tool } from './grid-engine';

@Component({
  selector: 'app-root',
  imports: [CommonModule, FormsModule],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  private readonly api = inject(EditorApiService);

  protected readonly loading = signal(true);
  protected readonly status = signal('');
  protected readonly bootstrap = signal<BootstrapResponse | null>(null);

  protected readonly collectionName = signal('');
  protected readonly spaceName = signal('');
  protected readonly areaName = signal('');

  protected readonly fixture = signal<'prototype' | 'fragment' | 'interactable' | 'transformation' | 'ground'>('prototype');
  protected readonly tool = signal<Tool>('select');
  protected readonly selectedAssetId = signal('');
  protected readonly selection = signal<GridSelection | undefined>(undefined);

  protected readonly prototypeSet = signal('');
  protected readonly fragmentSet = signal('');
  protected readonly interactableSet = signal('');

  protected readonly collectionNames = computed(() => Object.keys(this.bootstrap()?.collections ?? {}));

  protected readonly currentCollection = computed<Collection | undefined>(() => {
    const all = this.bootstrap()?.collections;
    if (!all) {
      return undefined;
    }
    return all[this.collectionName()];
  });

  protected readonly spaceNames = computed(() => Object.keys(this.currentCollection()?.Spaces ?? {}));

  protected readonly currentSpace = computed<Space | undefined>(() => {
    const spaces = this.currentCollection()?.Spaces;
    if (!spaces) {
      return undefined;
    }
    return spaces[this.spaceName()];
  });

  protected readonly areaNames = computed(() => this.currentSpace()?.Areas.map((area) => area.Name) ?? []);

  protected readonly currentArea = computed<AreaDescription | undefined>(() => {
    return this.currentSpace()?.Areas.find((area) => area.Name === this.areaName());
  });

  protected readonly prototypeSets = computed(() => Object.keys(this.currentCollection()?.PrototypeSets ?? {}));
  protected readonly fragmentSets = computed(() => Object.keys(this.currentCollection()?.Fragments ?? {}));
  protected readonly interactableSets = computed(() => Object.keys(this.currentCollection()?.InteractableSets ?? {}));

  protected readonly prototypes = computed<Prototype[]>(() => {
    const setName = this.prototypeSet();
    return this.currentCollection()?.PrototypeSets[setName] ?? [];
  });

  protected readonly fragments = computed<Fragment[]>(() => {
    const setName = this.fragmentSet();
    return this.currentCollection()?.Fragments[setName] ?? [];
  });

  protected readonly interactables = computed<InteractableDescription[]>(() => {
    const setName = this.interactableSet();
    return this.currentCollection()?.InteractableSets[setName] ?? [];
  });

  protected readonly prototypesById = computed(() => {
    const map = new Map<string, Prototype>();
    const sets = this.currentCollection()?.PrototypeSets ?? {};
    for (const setName of Object.keys(sets)) {
      for (const prototype of sets[setName]) {
        map.set(prototype.id, prototype);
      }
    }
    return map;
  });

  protected readonly fragmentsById = computed(() => {
    const map = new Map<string, Fragment>();
    const sets = this.currentCollection()?.Fragments ?? {};
    for (const setName of Object.keys(sets)) {
      for (const fragment of sets[setName]) {
        map.set(fragment.id, fragment);
      }
    }
    return map;
  });

  protected readonly interactablesById = computed(() => {
    const map = new Map<string, InteractableDescription>();
    const sets = this.currentCollection()?.InteractableSets ?? {};
    for (const setName of Object.keys(sets)) {
      for (const interactable of sets[setName]) {
        map.set(interactable.id, interactable);
      }
    }
    return map;
  });

  protected readonly gridMaterials = computed<Material[][]>(() => {
    const area = this.currentArea();
    if (!area) {
      return [];
    }
    return generateMaterials(area.Blueprint, this.prototypesById(), this.fixture() === 'ground');
  });

  protected readonly gridInteractables = computed(() => {
    if (this.fixture() === 'ground') {
      return [] as (InteractableDescription | undefined)[][];
    }
    const area = this.currentArea();
    if (!area) {
      return [] as (InteractableDescription | undefined)[][];
    }
    return area.Blueprint.Tiles.map((row) => row.map((tile) => this.interactablesById().get(tile.interactableId ?? '')));
  });

  constructor() {
    this.ensureLegacyStyles();
    void this.loadBootstrap();
  }

  private ensureLegacyStyles(): void {
    const stylesheets = ['/assets/style.css', '/assets/colors.css', '/assets/form.css'];
    for (const href of stylesheets) {
      const alreadyPresent = Array.from(document.querySelectorAll('link[rel="stylesheet"]')).some(
        (el) => (el as HTMLLinkElement).href.endsWith(href),
      );
      if (alreadyPresent) {
        continue;
      }
      const link = document.createElement('link');
      link.rel = 'stylesheet';
      link.href = href;
      document.head.appendChild(link);
    }
  }

  protected getEffectiveTool(): Tool {
    if (this.fixture() === 'ground') {
      if (
        this.tool() === 'select' ||
        this.tool() === 'toggle' ||
        this.tool() === 'toggle-fill' ||
        this.tool() === 'toggle-between'
      ) {
        return this.tool();
      }
      return 'toggle';
    }
    return this.tool();
  }

  protected onCollectionChange(): void {
    const firstSpace = this.spaceNames()[0] ?? '';
    this.spaceName.set(firstSpace);
    this.onSpaceChange();

    this.prototypeSet.set(this.prototypeSets()[0] ?? '');
    this.fragmentSet.set(this.fragmentSets()[0] ?? '');
    this.interactableSet.set(this.interactableSets()[0] ?? '');
    this.selectedAssetId.set('');
  }

  protected onSpaceChange(): void {
    this.areaName.set(this.areaNames()[0] ?? '');
    this.selection.set(undefined);
  }

  protected onFixtureChange(): void {
    this.selection.set(undefined);
    switch (this.fixture()) {
      case 'prototype':
        this.tool.set('select');
        this.selectedAssetId.set(this.prototypes()[0]?.id ?? '');
        break;
      case 'fragment':
        this.tool.set('place');
        this.selectedAssetId.set(this.fragments()[0]?.id ?? '');
        break;
      case 'interactable':
        this.tool.set('interactable-replace');
        this.selectedAssetId.set(this.interactables()[0]?.id ?? '');
        break;
      case 'transformation':
        this.tool.set('rotate');
        this.selectedAssetId.set('');
        break;
      case 'ground':
        this.tool.set('toggle');
        this.selectedAssetId.set('');
        break;
    }
  }

  protected onPrototypeSetChange(): void {
    this.selectedAssetId.set(this.prototypes()[0]?.id ?? '');
  }

  protected onFragmentSetChange(): void {
    this.selectedAssetId.set(this.fragments()[0]?.id ?? '');
  }

  protected onInteractableSetChange(): void {
    this.selectedAssetId.set(this.interactables()[0]?.id ?? '');
  }

  protected onGridClick(y: number, x: number): void {
    const area = this.currentArea();
    if (!area) {
      return;
    }

    const nextSelection = applyGridTool({
      y,
      x,
      tool: this.getEffectiveTool(),
      selectedAssetId: this.selectedAssetId(),
      selected: this.selection(),
      blueprint: area.Blueprint,
      prototypesById: this.prototypesById(),
      fragmentsById: this.fragmentsById(),
      interactablesById: this.interactablesById(),
    });

    this.selection.set(nextSelection);
    this.bootstrap.set({ ...(this.bootstrap() as BootstrapResponse) });
  }

  protected async saveSpace(): Promise<void> {
    const colName = this.collectionName();
    const sName = this.spaceName();
    const space = this.currentSpace();
    if (!colName || !sName || !space) {
      return;
    }

    this.status.set('Saving...');
    await this.api.saveSpace(colName, sName, space);
    this.status.set('Saved.');
  }

  protected async compileCollection(): Promise<void> {
    const colName = this.collectionName();
    if (!colName) {
      return;
    }
    this.status.set('Compiling...');
    await this.api.compile(colName);
    this.status.set('Compiled.');
  }

  protected async deployCollection(): Promise<void> {
    const colName = this.collectionName();
    if (!colName) {
      return;
    }
    this.status.set('Deploying...');
    await this.api.deploy(colName);
    this.status.set('Deployed.');
  }

  private async loadBootstrap(): Promise<void> {
    this.loading.set(true);
    const raw = await this.api.getBootstrap();
    const data = this.normalizeBootstrap(raw as unknown as Record<string, unknown>);
    this.bootstrap.set(data);

    this.collectionName.set(this.collectionNames()[0] ?? '');
    this.spaceName.set(this.spaceNames()[0] ?? '');
    this.areaName.set(this.areaNames()[0] ?? '');

    this.prototypeSet.set(this.prototypeSets()[0] ?? '');
    this.fragmentSet.set(this.fragmentSets()[0] ?? '');
    this.interactableSet.set(this.interactableSets()[0] ?? '');

    this.selectedAssetId.set(this.prototypes()[0]?.id ?? '');
    this.tool.set('select');
    this.loading.set(false);
  }

  private normalizeBootstrap(raw: Record<string, unknown>): BootstrapResponse {
    const sourceCollections = (raw['collections'] ?? {}) as Record<string, any>;
    const outCollections: Record<string, Collection> = {};

    for (const collectionName of Object.keys(sourceCollections)) {
      const collection = sourceCollections[collectionName] ?? {};
      const spaces: Record<string, Space> = {};
      const sourceSpaces = (collection.Spaces ?? collection.spaces ?? {}) as Record<string, any>;
      for (const spaceName of Object.keys(sourceSpaces)) {
        const s = sourceSpaces[spaceName];
        spaces[spaceName] = {
          CollectionName: s.CollectionName ?? s.collectionName ?? collectionName,
          Name: s.Name ?? s.name ?? spaceName,
          Topology: s.Topology ?? s.topology ?? '',
          Latitude: s.Latitude ?? s.latitude ?? 0,
          Longitude: s.Longitude ?? s.longitude ?? 0,
          AreaHeight: s.AreaHeight ?? s.areaHeight ?? 0,
          AreaWidth: s.AreaWidth ?? s.areaWidth ?? 0,
          Areas: (s.Areas ?? s.areas ?? []).map((area: any) => ({
            Name: area.Name ?? area.name ?? '',
            Safe: area.Safe ?? area.safe ?? false,
            Blueprint: {
              Tiles: area.Blueprint?.Tiles ?? area.Blueprint?.tiles ?? area.blueprint?.Tiles ?? area.blueprint?.tiles ?? [],
              Instructions:
                area.Blueprint?.Instructions ??
                area.Blueprint?.instructions ??
                area.blueprint?.Instructions ??
                area.blueprint?.instructions ??
                [],
              Ground: area.Blueprint?.Ground ?? area.blueprint?.Ground ?? area.blueprint?.ground,
              DefaultTileColor:
                area.Blueprint?.DefaultTileColor ??
                area.Blueprint?.defaultTileColor ??
                area.blueprint?.DefaultTileColor ??
                area.blueprint?.defaultTileColor ??
                'white',
              DefaultTileColor1:
                area.Blueprint?.DefaultTileColor1 ??
                area.Blueprint?.defaultTileColor1 ??
                area.blueprint?.DefaultTileColor1 ??
                area.blueprint?.defaultTileColor1 ??
                'white',
            },
            Transports: area.Transports ?? area.transports ?? [],
            North: area.North ?? area.north,
            South: area.South ?? area.south,
            East: area.East ?? area.east,
            West: area.West ?? area.west,
            Weather: area.Weather ?? area.weather,
            LoadStrategy: area.LoadStrategy ?? area.loadStrategy,
            SpawnStrategy: area.SpawnStrategy ?? area.spawnStrategy,
            BroadcastGroup: area.BroadcastGroup ?? area.broadcastGroup,
          })),
        };
      }

      outCollections[collectionName] = {
        Name: collection.Name ?? collection.name ?? collectionName,
        Spaces: spaces,
        Fragments: collection.Fragments ?? collection.fragments ?? {},
        PrototypeSets: collection.PrototypeSets ?? collection.prototypeSets ?? {},
        InteractableSets: collection.InteractableSets ?? collection.interactableSets ?? {},
        StructureSets: collection.StructureSets ?? collection.structureSets ?? {},
      };
    }

    return {
      collections: outCollections,
      colors: (raw['colors'] ?? []) as BootstrapResponse['colors'],
    };
  }
}
