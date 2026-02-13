import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { EditorApiService } from './editor-api.service';
import {
  AreaDescription,
  Blueprint,
  BootstrapResponse,
  Collection,
  Color,
  Fragment,
  GridSelection,
  InteractableDescription,
  Material,
  Prototype,
  Space,
} from './editor.models';
import { applyGridTool, generateMaterials, Tool } from './grid-engine';

type ViewMode = 'world' | 'create' | 'prototypes' | 'fragments' | 'interactables' | 'colors';
type GridTarget = 'area' | 'fragment';

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
  protected readonly gridVersion = signal(0);

  protected readonly viewMode = signal<ViewMode>('world');
  protected readonly gridTarget = signal<GridTarget>('area');
  protected readonly createTarget = signal<'space' | 'area'>('space');

  protected readonly collectionName = signal('');
  protected readonly showNewCollection = signal(false);
  protected readonly spaceName = signal('');
  protected readonly areaName = signal('');

  protected readonly fixture = signal<'prototype' | 'fragment' | 'interactable' | 'transformation' | 'ground'>('prototype');
  protected readonly tool = signal<Tool>('select');
  protected readonly selectedAssetId = signal('');
  protected readonly selection = signal<GridSelection | undefined>(undefined);

  protected readonly prototypeSet = signal('');
  protected readonly fragmentSet = signal('');
  protected readonly interactableSet = signal('');

  protected readonly prototypeEditId = signal('');
  protected readonly fragmentEditId = signal('');
  protected readonly interactableEditId = signal('');
  protected readonly colorEditIndex = signal(0);

  protected readonly newCollectionName = signal('');
  protected readonly newSpace = signal({
    name: '',
    topology: 'plane',
    latitude: 8,
    longitude: 8,
    areaWidth: 16,
    areaHeight: 16,
    tileColor: 'green',
    tileColor1: 'brown',
    weather: '',
    broadcastGroup: '',
  });
  protected readonly newArea = signal({
    name: '',
    safe: false,
    height: 16,
    width: 16,
    defaultTileColor: 'green',
    defaultTileColor1: 'brown',
  });

  protected readonly newPrototypeSetName = signal('');
  protected readonly newFragmentSetName = signal('');
  protected readonly newInteractableSetName = signal('');
  protected readonly newColor = signal({ cssClassName: '', R: 0, G: 0, B: 0, A: '' });

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

  protected readonly editedPrototype = computed<Prototype | undefined>(() => {
    return this.prototypes().find((proto) => proto.id === this.prototypeEditId());
  });

  protected readonly editedFragment = computed<Fragment | undefined>(() => {
    return this.fragments().find((fragment) => fragment.id === this.fragmentEditId());
  });

  protected readonly editedInteractable = computed<InteractableDescription | undefined>(() => {
    return this.interactables().find((entry) => entry.id === this.interactableEditId());
  });

  protected readonly colors = computed<Color[]>(() => this.bootstrap()?.colors ?? []);

  protected readonly editedColor = computed<Color | undefined>(() => {
    const idx = this.colorEditIndex();
    return this.colors()[idx];
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

  protected readonly activeBlueprint = computed<Blueprint | undefined>(() => {
    if (this.gridTarget() === 'fragment') {
      return this.editedFragment()?.blueprint;
    }
    return this.currentArea()?.Blueprint;
  });

  protected readonly gridMaterials = computed<Material[][]>(() => {
    this.gridVersion();
    const blueprint = this.activeBlueprint();
    if (!blueprint) {
      return [];
    }
    return generateMaterials(blueprint, this.prototypesById(), this.fixture() === 'ground');
  });

  protected readonly gridInteractables = computed(() => {
    this.gridVersion();
    if (this.fixture() === 'ground') {
      return [] as (InteractableDescription | undefined)[][];
    }
    const blueprint = this.activeBlueprint();
    if (!blueprint) {
      return [] as (InteractableDescription | undefined)[][];
    }
    return blueprint.Tiles.map((row) => row.map((tile) => this.interactablesById().get(tile.interactableId ?? '')));
  });

  constructor() {
    this.ensureLegacyStyles();
    void this.loadBootstrap();
  }

  protected setViewMode(mode: ViewMode): void {
    this.viewMode.set(mode);
    this.gridTarget.set(mode === 'fragments' ? 'fragment' : 'area');
    this.selection.set(undefined);
  }

  protected setCreateTarget(target: 'space' | 'area'): void {
    this.createTarget.set(target);
  }

  protected onCollectionChange(): void {
    this.spaceName.set(this.spaceNames()[0] ?? '');
    this.onSpaceChange();

    this.prototypeSet.set(this.prototypeSets()[0] ?? '');
    this.fragmentSet.set(this.fragmentSets()[0] ?? '');
    this.interactableSet.set(this.interactableSets()[0] ?? '');

    this.onPrototypeSetChange();
    this.onFragmentSetChange();
    this.onInteractableSetChange();
  }

  protected onSpaceChange(): void {
    this.areaName.set(this.areaNames()[0] ?? '');
    const area = this.currentArea();
    if (area) {
      this.newArea.set({
        ...this.newArea(),
        defaultTileColor: area.Blueprint.DefaultTileColor || 'green',
        defaultTileColor1: area.Blueprint.DefaultTileColor1 || 'brown',
      });
    }
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
    const first = this.prototypes()[0];
    this.selectedAssetId.set(first?.id ?? '');
    this.prototypeEditId.set(first?.id ?? '');
  }

  protected onFragmentSetChange(): void {
    const first = this.fragments()[0];
    this.selectedAssetId.set(first?.id ?? '');
    this.fragmentEditId.set(first?.id ?? '');
  }

  protected onInteractableSetChange(): void {
    const first = this.interactables()[0];
    this.selectedAssetId.set(first?.id ?? '');
    this.interactableEditId.set(first?.id ?? '');
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

  protected onGridClick(y: number, x: number): void {
    const blueprint = this.activeBlueprint();
    if (!blueprint) {
      return;
    }

    const nextSelection = applyGridTool({
      y,
      x,
      tool: this.getEffectiveTool(),
      selectedAssetId: this.selectedAssetId(),
      selected: this.selection(),
      blueprint,
      prototypesById: this.prototypesById(),
      fragmentsById: this.fragmentsById(),
      interactablesById: this.interactablesById(),
    });

    this.selection.set(nextSelection);
    this.touchBootstrap();
  }

  protected addInstruction(): void {
    const blueprint = this.activeBlueprint();
    if (!blueprint || !this.selectedAssetId()) {
      return;
    }
    blueprint.Instructions = blueprint.Instructions ?? [];
    blueprint.Instructions.push({
      ID: crypto.randomUUID(),
      X: this.selection()?.x ?? 0,
      Y: this.selection()?.y ?? 0,
      GridAssetId: this.selectedAssetId(),
      ClockwiseRotations: 0,
    });
    this.touchBootstrap();
  }

  protected moveInstruction(index: number): void {
    const blueprint = this.activeBlueprint();
    if (!blueprint || !blueprint.Instructions || blueprint.Instructions.length < 2) {
      return;
    }
    const next = (index + 1) % blueprint.Instructions.length;
    const hold = blueprint.Instructions[index];
    blueprint.Instructions[index] = blueprint.Instructions[next];
    blueprint.Instructions[next] = hold;
    this.touchBootstrap();
  }

  protected deleteInstruction(index: number): void {
    const blueprint = this.activeBlueprint();
    if (!blueprint) {
      return;
    }
    blueprint.Instructions.splice(index, 1);
    this.touchBootstrap();
  }

  protected showInstruction(index: number): void {
    const blueprint = this.activeBlueprint();
    if (!blueprint?.Instructions[index]) {
      return;
    }
    const instruction = blueprint.Instructions[index];
    this.selection.set({ y: instruction.Y, x: instruction.X });
  }

  protected async createCollection(): Promise<void> {
    if (!this.newCollectionName().trim()) {
      return;
    }
    this.status.set('Creating collection...');
    await this.api.createCollection(this.newCollectionName().trim());
    this.newCollectionName.set('');
    await this.loadBootstrap();
    this.showNewCollection.set(false);
    this.status.set('Collection created.');
  }

  protected toggleNewCollectionForm(): void {
    this.showNewCollection.update((value) => !value);
  }

  protected hideNewCollectionForm(): void {
    this.showNewCollection.set(false);
  }

  protected async createSpace(): Promise<void> {
    const colName = this.collectionName();
    if (!colName || !this.newSpace().name.trim()) {
      return;
    }

    this.status.set('Creating space...');
    await this.api.createSpace({
      collectionName: colName,
      ...this.newSpace(),
      name: this.newSpace().name.trim(),
    });
    this.newSpace.set({ ...this.newSpace(), name: '' });
    await this.loadBootstrap();
    this.status.set('Space created.');
  }

  protected async createArea(): Promise<void> {
    const colName = this.collectionName();
    const sName = this.spaceName();
    if (!colName || !sName || !this.newArea().name.trim()) {
      return;
    }

    this.status.set('Creating area...');
    await this.api.createArea({
      collectionName: colName,
      spaceName: sName,
      ...this.newArea(),
      name: this.newArea().name.trim(),
    });
    this.newArea.set({ ...this.newArea(), name: '' });
    await this.loadBootstrap();
    this.status.set('Area created.');
  }

  protected addPrototypeSet(): void {
    const name = this.newPrototypeSetName().trim();
    const collection = this.currentCollection();
    if (!collection || !name || collection.PrototypeSets[name]) {
      return;
    }
    collection.PrototypeSets[name] = [];
    this.prototypeSet.set(name);
    this.newPrototypeSetName.set('');
    this.touchBootstrap();
  }

  protected addPrototype(): void {
    const setName = this.prototypeSet();
    const collection = this.currentCollection();
    if (!collection || !setName) {
      return;
    }
    const next: Prototype = {
      id: crypto.randomUUID(),
      setName,
      commonName: 'new-prototype',
      cssColor: '',
      walkable: false,
      layer1css: '',
      layer2css: '',
      ceiling1css: '',
      ceiling2css: '',
      mapColor: '',
      editorColor: '',
      displayText: '',
    };
    collection.PrototypeSets[setName].push(next);
    this.prototypeEditId.set(next.id);
    this.selectedAssetId.set(next.id);
    this.touchBootstrap();
  }

  protected async savePrototypeSet(): Promise<void> {
    const colName = this.collectionName();
    const setName = this.prototypeSet();
    if (!colName || !setName) {
      return;
    }
    this.status.set('Saving prototype set...');
    await this.api.savePrototypeSet(colName, setName, this.prototypes());
    this.status.set('Prototype set saved.');
  }

  protected addFragmentSet(): void {
    const name = this.newFragmentSetName().trim();
    const collection = this.currentCollection();
    if (!collection || !name || collection.Fragments[name]) {
      return;
    }
    collection.Fragments[name] = [];
    this.fragmentSet.set(name);
    this.newFragmentSetName.set('');
    this.touchBootstrap();
  }

  protected addFragment(): void {
    const setName = this.fragmentSet();
    const collection = this.currentCollection();
    if (!collection || !setName) {
      return;
    }
    const area = this.currentArea();
    const h = area?.Blueprint.Tiles.length ?? 8;
    const w = area?.Blueprint.Tiles[0]?.length ?? 8;
    const tiles = Array.from({ length: h }, () => Array.from({ length: w }, () => ({ prototypeId: '', interactableId: '' })));

    const next: Fragment = {
      id: crypto.randomUUID(),
      setName,
      name: `fragment-${collection.Fragments[setName].length + 1}`,
      blueprint: {
        Tiles: tiles,
        Instructions: [],
        DefaultTileColor: area?.Blueprint.DefaultTileColor ?? 'green',
        DefaultTileColor1: area?.Blueprint.DefaultTileColor1 ?? 'brown',
      },
    };
    collection.Fragments[setName].push(next);
    this.fragmentEditId.set(next.id);
    this.selectedAssetId.set(next.id);
    this.touchBootstrap();
  }

  protected async saveFragmentSet(): Promise<void> {
    const colName = this.collectionName();
    const setName = this.fragmentSet();
    if (!colName || !setName) {
      return;
    }
    this.status.set('Saving fragment set...');
    await this.api.saveFragmentSet(colName, setName, this.fragments());
    this.status.set('Fragment set saved.');
  }

  protected addInteractableSet(): void {
    const name = this.newInteractableSetName().trim();
    const collection = this.currentCollection();
    if (!collection || !name || collection.InteractableSets[name]) {
      return;
    }
    collection.InteractableSets[name] = [];
    this.interactableSet.set(name);
    this.newInteractableSetName.set('');
    this.touchBootstrap();
  }

  protected addInteractable(): void {
    const setName = this.interactableSet();
    const collection = this.currentCollection();
    if (!collection || !setName) {
      return;
    }
    const next: InteractableDescription = {
      id: crypto.randomUUID(),
      setName,
      name: 'new-interactable',
      cssClass: '',
      pushable: false,
      walkable: false,
      fragile: false,
      reactions: '',
    };
    collection.InteractableSets[setName].push(next);
    this.interactableEditId.set(next.id);
    this.selectedAssetId.set(next.id);
    this.touchBootstrap();
  }

  protected async saveInteractableSet(): Promise<void> {
    const colName = this.collectionName();
    const setName = this.interactableSet();
    if (!colName || !setName) {
      return;
    }
    this.status.set('Saving interactable set...');
    await this.api.saveInteractableSet(colName, setName, this.interactables());
    this.status.set('Interactable set saved.');
  }

  protected addColor(): void {
    const model = this.newColor();
    if (!model.cssClassName.trim()) {
      return;
    }
    const next: Color = {
      cssClassName: model.cssClassName.trim(),
      R: Number(model.R) || 0,
      G: Number(model.G) || 0,
      B: Number(model.B) || 0,
      A: model.A ?? '',
    };
    const colors = this.colors();
    colors.push(next);
    this.colorEditIndex.set(colors.length - 1);
    this.newColor.set({ cssClassName: '', R: 0, G: 0, B: 0, A: '' });
    this.touchBootstrap();
  }

  protected async saveColors(): Promise<void> {
    this.status.set('Saving colors...');
    await this.api.saveColors(this.colors());
    this.status.set('Colors saved.');
  }

  protected async saveSpace(): Promise<void> {
    const colName = this.collectionName();
    const sName = this.spaceName();
    const space = this.currentSpace();
    if (!colName || !sName || !space) {
      return;
    }

    this.status.set('Saving space...');
    await this.api.saveSpace(colName, sName, space);
    this.status.set('Space saved.');
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

  private ensureLegacyStyles(): void {
    const stylesheets = ['/assets/style.css', '/assets/colors.css'];
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

  private touchBootstrap(): void {
    this.gridVersion.update((value) => value + 1);
    this.bootstrap.set({ ...(this.bootstrap() as BootstrapResponse) });
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

    this.onPrototypeSetChange();
    this.onFragmentSetChange();
    this.onInteractableSetChange();

    this.selectedAssetId.set(this.prototypes()[0]?.id ?? '');
    this.tool.set('select');
    this.loading.set(false);
  }

  private normalizeBlueprint(input: any): Blueprint {
    return {
      Tiles: input?.Tiles ?? input?.tiles ?? [],
      Instructions: input?.Instructions ?? input?.instructions ?? [],
      Ground: input?.Ground ?? input?.ground,
      DefaultTileColor: input?.DefaultTileColor ?? input?.defaultTileColor ?? 'white',
      DefaultTileColor1: input?.DefaultTileColor1 ?? input?.defaultTileColor1 ?? 'white',
    };
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
            Blueprint: this.normalizeBlueprint(area.Blueprint ?? area.blueprint ?? {}),
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

      const outFragments: Record<string, Fragment[]> = {};
      const sourceFragments = (collection.Fragments ?? collection.fragments ?? {}) as Record<string, any[]>;
      for (const setName of Object.keys(sourceFragments)) {
        outFragments[setName] = (sourceFragments[setName] ?? []).map((fragment) => ({
          id: fragment.id ?? fragment.ID ?? crypto.randomUUID(),
          name: fragment.name ?? fragment.Name ?? 'fragment',
          setName: fragment.setName ?? fragment.SetName ?? setName,
          blueprint: this.normalizeBlueprint(fragment.blueprint ?? fragment.Blueprint ?? {}),
        }));
      }

      outCollections[collectionName] = {
        Name: collection.Name ?? collection.name ?? collectionName,
        Spaces: spaces,
        Fragments: outFragments,
        PrototypeSets: (collection.PrototypeSets ?? collection.prototypeSets ?? {}) as Collection['PrototypeSets'],
        InteractableSets: (collection.InteractableSets ?? collection.interactableSets ?? {}) as Collection['InteractableSets'],
        StructureSets: collection.StructureSets ?? collection.structureSets ?? {},
      };
    }

    return {
      collections: outCollections,
      colors: (raw['colors'] ?? []) as BootstrapResponse['colors'],
    };
  }
}
