import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { EditorApiService } from '../../core/services/editor-api.service';
import {
  AreaDescription,
  Blueprint,
  BootstrapResponse,
  Cell,
  Collection,
  Color,
  Fragment,
  GridSelection,
  InteractableDescription,
  InteractableStateDescription,
  Instruction,
  Material,
  Prototype,
  ReactionRule,
  Space,
  Transport,
} from '../../core/models/editor.models';
import {
  REACTS_WITH_REGISTRY,
  REACTION_REGISTRY,
  findReactsWithEntry,
  findReactionEntry,
  RegistryEntry,
} from '../../core/models/reaction-registry';
import {
  applyGridTool,
  deleteInstructionAndReapply,
  generateMaterials,
  normalizeInstructionField,
  reorderInstructionAndReapply,
  Tool,
  updateInstructionAndReapply,
} from './grid-engine';

type ViewMode = 'world' | 'create' | 'modify-space' | 'prototypes' | 'fragments' | 'interactables' | 'colors';
type GridTarget = 'area' | 'fragment';
type BulkAreaProperty =
  | 'safe'
  | 'defaultTileColor'
  | 'defaultTileColor1'
  | 'weather'
  | 'loadStrategy'
  | 'spawnStrategy'
  | 'broadcastGroup';

interface NavigationMapCell {
  areaName: string;
  row: number;
  column: number;
  imageUrl: string;
  exists: boolean;
  isCurrent: boolean;
}

type ResolvedAssetType = 'prototype' | 'fragment' | 'interactable';

interface SelectedTileInstructionInfo {
  instruction: Instruction;
  index: number;
  assetType?: ResolvedAssetType;
  assetLabel: string;
}

@Component({
  selector: 'app-editor',
  imports: [CommonModule, FormsModule],
  templateUrl: './editor.component.html',
  styleUrl: './editor.component.css',
})
export class EditorComponent {
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
  protected readonly modifySpaceName = signal('');
  protected readonly areaName = signal('');
  protected readonly showGridTools = signal(true);
  protected readonly showAreaDetails = signal(false);
  protected readonly showTransports = signal(false);
  protected readonly showNeighbors = signal(false);
  protected readonly showNavigationMap = signal(false);
  protected readonly showBlueprintInstructions = signal(false);
  protected readonly showSelectedInformation = signal(false);
  protected readonly instructionEditedIds = signal<Record<string, true>>({});
  protected readonly sideColumnWidth = signal(360);
  protected readonly isResizingSideColumn = signal(false);
  protected readonly failedNavigationImageKeys = signal<Record<string, true>>({});

  protected readonly fixture = signal<'prototype' | 'fragment' | 'interactable' | 'transformation' | 'ground'>('prototype');
  protected readonly tool = signal<Tool>('select');
  protected readonly selectedAssetId = signal('');
  protected readonly selection = signal<GridSelection | undefined>(undefined);
  protected readonly hoverPosition = signal<GridSelection | undefined>(undefined);

  protected readonly prototypeSet = signal('');
  protected readonly fragmentSet = signal('');
  protected readonly interactableSet = signal('');
  protected readonly interactableStateEditName = signal('default');
  protected readonly selectedInteractableState = signal('default');

  protected readonly prototypeEditId = signal('');
  protected readonly fragmentEditId = signal('');
  protected readonly interactableEditId = signal('');
  protected readonly bulkAreaProperty = signal<BulkAreaProperty>('safe');
  protected readonly bulkAreaValueText = signal('');
  protected readonly bulkAreaValueBoolean = signal(false);
  protected readonly colorEditIndex = signal(0);
  protected readonly prototypePreviewUseEditorColor = signal(false);

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

  // Reaction rule registries exposed to the template
  protected readonly reactsWithRegistry = REACTS_WITH_REGISTRY;
  protected readonly reactionRegistry = REACTION_REGISTRY;

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

  protected readonly canFlattenCurrentSpace = computed(() => {
    const space = this.currentSpace();
    return Boolean(space && this.isSimplyTiledSpace(space));
  });

  protected readonly modifyTargetSpace = computed<Space | undefined>(() => {
    const spaces = this.currentCollection()?.Spaces;
    if (!spaces) {
      return undefined;
    }
    return spaces[this.modifySpaceName()];
  });

  protected readonly modifyTargetAreaCount = computed(() => this.modifyTargetSpace()?.Areas.length ?? 0);

  protected readonly areaNames = computed(() => this.currentSpace()?.Areas.map((area) => area.Name) ?? []);

  protected readonly currentArea = computed<AreaDescription | undefined>(() => {
    return this.currentSpace()?.Areas.find((area) => area.Name === this.areaName());
  });

  protected readonly transportSourceKeys = computed(() => {
    const area = this.currentArea();
    const keys = new Set<string>();
    if (!area) {
      return keys;
    }

    for (const transport of area.Transports ?? []) {
      const y = Number(transport.SourceY);
      const x = Number(transport.SourceX);
      if (Number.isInteger(y) && Number.isInteger(x)) {
        keys.add(`${y}:${x}`);
      }
    }

    return keys;
  });

  protected isTransportSourceTile(y: number, x: number): boolean {
    return this.transportSourceKeys().has(`${y}:${x}`);
  }

  protected readonly hasNavigationMap = computed(() => {
    const space = this.currentSpace();
    if (!space) {
      return false;
    }
    return this.isSimplyTiledSpace(space) && space.Latitude > 0 && space.Longitude > 0;
  });

  protected readonly navigationMapRows = computed<NavigationMapCell[][]>(() => {
    const space = this.currentSpace();
    const collectionName = this.collectionName();
    if (!space || !this.hasNavigationMap()) {
      return [];
    }

    const areaNames = new Set(space.Areas.map((area) => area.Name));
    const selected = this.areaName();
    const rows: NavigationMapCell[][] = [];

    for (let row = 0; row < space.Latitude; row += 1) {
      const mapRow: NavigationMapCell[] = [];
      for (let column = 0; column < space.Longitude; column += 1) {
        const areaName = `${space.Name}:${row}-${column}`;
        const exists = areaNames.has(areaName);
        mapRow.push({
          areaName,
          row,
          column,
          imageUrl: this.buildAreaImageUrl(space.Name, areaName, collectionName),
          exists,
          isCurrent: selected === areaName,
        });
      }
      rows.push(mapRow);
    }

    return rows;
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

  protected readonly editedInteractableStateNames = computed<string[]>(() => {
    const interactable = this.editedInteractable();
    if (!interactable) {
      return [];
    }
    this.ensureInteractableStateModel(interactable);
    return Object.keys(interactable.states ?? {});
  });

  protected readonly editedInteractableState = computed<InteractableStateDescription | undefined>(() => {
    const interactable = this.editedInteractable();
    if (!interactable) {
      return undefined;
    }
    this.ensureInteractableStateModel(interactable);
    const stateName = this.interactableStateEditName() || interactable.defaultState || 'default';
    return interactable.states?.[stateName];
  });

  protected readonly colors = computed<Color[]>(() => this.bootstrap()?.colors ?? []);

  protected readonly editedColor = computed<Color | undefined>(() => {
    const idx = this.colorEditIndex();
    return this.colors()[idx];
  });

  protected colorPreviewStyle(color: Color | undefined): Record<string, string> {
    if (!color) {
      return { background: 'transparent' };
    }
    const r = this.clampColorChannel(color.R);
    const g = this.clampColorChannel(color.G);
    const b = this.clampColorChannel(color.B);
    const alpha = this.clampAlpha(color.A);
    return { background: `rgba(${r}, ${g}, ${b}, ${alpha})` };
  }

  protected togglePrototypePreviewEditorColor(): void {
    this.prototypePreviewUseEditorColor.update((value) => !value);
  }

  protected prototypePreviewCeiling2Class(prototype: Prototype | undefined): string {
    if (!prototype) {
      return '';
    }
    const editorColor = prototype.editorColor?.trim() ?? '';
    if (this.prototypePreviewUseEditorColor() && editorColor) {
      return this.prototypePreviewClass(editorColor);
    }
    return this.prototypePreviewClass(prototype.ceiling2css);
  }

  protected prototypePreviewClass(value: string | undefined): string {
    const className = (value ?? '').trim();
    if (!className) {
      return '';
    }
    return className.replace(/\{[a-zA-Z0-9_-]+:([^}]+)\}/g, '$1').trim();
  }

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

  protected readonly selectedTile = computed(() => {
    const blueprint = this.activeBlueprint();
    const selection = this.selection();
    if (!blueprint || !selection) {
      return undefined;
    }
    const row = blueprint.Tiles[selection.y];
    if (!row || selection.x < 0 || selection.x >= row.length) {
      return undefined;
    }
    return row[selection.x];
  });

  protected readonly selectedTilePrototype = computed<Prototype | undefined>(() => {
    const prototypeId = this.selectedTile()?.prototypeId?.trim() ?? '';
    if (!prototypeId) {
      return undefined;
    }
    return this.prototypesById().get(prototypeId);
  });

  protected readonly selectedTileInteractable = computed<InteractableDescription | undefined>(() => {
    const tile = this.selectedTile();
    const interactableId = tile?.interactableId?.trim() ?? '';
    if (!interactableId || !tile) {
      return undefined;
    }
    return this.resolveInteractableForTile(tile);
  });

  protected readonly selectedTileInstructions = computed<SelectedTileInstructionInfo[]>(() => {
    const blueprint = this.activeBlueprint();
    const selection = this.selection();
    if (!blueprint || !selection) {
      return [];
    }
    return (blueprint.Instructions ?? [])
      .map((instruction, index) => ({ instruction, index }))
      .filter(({ instruction }) => instruction.Y === selection.y && instruction.X === selection.x)
      .map(({ instruction, index }) => {
        const resolved = this.resolveAssetById(instruction.GridAssetId);
        if (!resolved) {
          return {
            instruction,
            index,
            assetLabel: `Unknown asset (${instruction.GridAssetId || 'empty'})`,
          };
        }

        return {
          instruction,
          index,
          assetType: resolved.type,
          assetLabel: this.assetDisplayName(resolved),
        };
      });
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
    return blueprint.Tiles.map((row) => row.map((tile) => this.resolveInteractableForTile(tile)));
  });

  protected readonly fixturePreviewMaterials = computed<Material[][]>(() => {
    const fixture = this.fixture();
    const defaults = this.activeBlueprint();
    const defaultColor0 = defaults?.DefaultTileColor ?? 'white';
    const defaultColor1 = defaults?.DefaultTileColor1 ?? 'white';

    if (fixture === 'prototype') {
      const selectedId = this.selectedAssetId();
      if (!selectedId) {
        return [];
      }
      return generateMaterials(
        {
          Tiles: [[{ prototypeId: selectedId, interactableId: '' }]],
          Instructions: [],
          DefaultTileColor: defaultColor0,
          DefaultTileColor1: defaultColor1,
        },
        this.prototypesById(),
        false,
      );
    }

    if (fixture === 'fragment') {
      const fragment = this.fragmentsById().get(this.selectedAssetId());
      if (!fragment) {
        return [];
      }
      return generateMaterials(fragment.blueprint, this.prototypesById(), false);
    }

    if (fixture === 'interactable') {
      return generateMaterials(
        {
          Tiles: [[{ prototypeId: '', interactableId: '' }]],
          Instructions: [],
          DefaultTileColor: defaultColor0,
          DefaultTileColor1: defaultColor1,
        },
        this.prototypesById(),
        true,
      );
    }

    return [];
  });

  protected readonly fixturePreviewInteractables = computed(() => {
    const fixture = this.fixture();

    if (fixture === 'fragment') {
      const fragment = this.fragmentsById().get(this.selectedAssetId());
      if (!fragment) {
        return [] as (InteractableDescription | undefined)[][];
      }
      return fragment.blueprint.Tiles.map((row) => row.map((tile) => this.interactablesById().get(tile.interactableId ?? '')));
    }

    if (fixture === 'interactable') {
      return [[this.resolveInteractableForAssetAndState(this.selectedAssetId(), this.selectedInteractableState())]];
    }

    return [] as (InteractableDescription | undefined)[][];
  });

  protected readonly ghostMaterials = computed(() => {
    this.gridVersion();
    const blueprint = this.activeBlueprint();
    const hover = this.hoverPosition();
    if (!blueprint || !hover) {
      return [] as (Material | undefined)[][];
    }

    const out: (Material | undefined)[][] = blueprint.Tiles.map((row) => row.map(() => undefined));
    const defaultColor0 = blueprint.DefaultTileColor ?? 'white';
    const defaultColor1 = blueprint.DefaultTileColor1 ?? 'white';

    if (this.getEffectiveTool() === 'select') {
      return out;
    }

    if (this.fixture() === 'prototype') {
      const selectedId = this.selectedAssetId();
      if (!selectedId) {
        return out;
      }
      const preview = generateMaterials(
        {
          Tiles: [[{ prototypeId: selectedId, interactableId: '' }]],
          Instructions: [],
          DefaultTileColor: defaultColor0,
          DefaultTileColor1: defaultColor1,
        },
        this.prototypesById(),
        false,
      )[0]?.[0];
      if (preview && hover.y >= 0 && hover.y < out.length && hover.x >= 0 && hover.x < out[hover.y].length) {
        out[hover.y][hover.x] = preview;
      }
      return out;
    }

    if (this.fixture() === 'fragment') {
      const fragment = this.fragmentsById().get(this.selectedAssetId());
      if (!fragment) {
        return out;
      }
      const ghostFragment = generateMaterials(fragment.blueprint, this.prototypesById(), false);
      for (let y = 0; y < ghostFragment.length; y += 1) {
        const targetY = hover.y + y;
        if (targetY < 0 || targetY >= out.length) {
          continue;
        }
        for (let x = 0; x < ghostFragment[y].length; x += 1) {
          const targetX = hover.x + x;
          if (targetX < 0 || targetX >= out[targetY].length) {
            continue;
          }
          out[targetY][targetX] = ghostFragment[y][x];
        }
      }
    }

    return out;
  });

  protected readonly ghostInteractables = computed(() => {
    this.gridVersion();
    const blueprint = this.activeBlueprint();
    const hover = this.hoverPosition();
    if (!blueprint || !hover) {
      return [] as (InteractableDescription | undefined)[][];
    }

    const out: (InteractableDescription | undefined)[][] = blueprint.Tiles.map((row) => row.map(() => undefined));

    if (this.getEffectiveTool() === 'select') {
      return out;
    }

    if (this.getEffectiveTool() === 'interactable-delete') {
      return out;
    }

    if (this.fixture() === 'interactable') {
      const interactable = this.resolveInteractableForAssetAndState(this.selectedAssetId(), this.selectedInteractableState());
      if (interactable && hover.y >= 0 && hover.y < out.length && hover.x >= 0 && hover.x < out[hover.y].length) {
        out[hover.y][hover.x] = interactable;
      }
      return out;
    }

    if (this.fixture() === 'fragment') {
      const fragment = this.fragmentsById().get(this.selectedAssetId());
      if (!fragment) {
        return out;
      }

      for (let y = 0; y < fragment.blueprint.Tiles.length; y += 1) {
        const targetY = hover.y + y;
        if (targetY < 0 || targetY >= out.length) {
          continue;
        }
        for (let x = 0; x < fragment.blueprint.Tiles[y].length; x += 1) {
          const targetX = hover.x + x;
          if (targetX < 0 || targetX >= out[targetY].length) {
            continue;
          }
          const interactableId = fragment.blueprint.Tiles[y][x].interactableId ?? '';
          out[targetY][targetX] = this.resolveInteractableForTile(fragment.blueprint.Tiles[y][x]);
        }
      }
    }

    return out;
  });

  protected isHoverTile(y: number, x: number): boolean {
    const hover = this.hoverPosition();
    return Boolean(hover && hover.y === y && hover.x === x);
  }

  protected shouldShowSelectHover(y: number, x: number): boolean {
    return this.getEffectiveTool() === 'select' && this.isHoverTile(y, x);
  }

  constructor() {
    this.ensureLegacyStyles();
    void this.loadBootstrap();
  }

  protected setViewMode(mode: ViewMode): void {
    this.viewMode.set(mode);
    this.gridTarget.set(mode === 'fragments' ? 'fragment' : 'area');
    this.selection.set(undefined);
    if (mode !== 'world') {
      this.resetAreaEditPanels();
    }
  }

  protected setCreateTarget(target: 'space' | 'area'): void {
    this.createTarget.set(target);
  }

  protected toggleAreaDetails(): void {
    const next = !this.showAreaDetails();
    if (next) {
      this.showGridTools.set(false);
      this.showTransports.set(false);
      this.showNeighbors.set(false);
    } else {
      this.showGridTools.set(true);
    }
    this.showAreaDetails.set(next);
  }

  protected toggleTransports(): void {
    const next = !this.showTransports();
    if (next) {
      this.showGridTools.set(false);
      this.showAreaDetails.set(false);
      this.showNeighbors.set(false);
    } else {
      this.showGridTools.set(true);
    }
    this.showTransports.set(next);
  }

  protected toggleNeighbors(): void {
    const next = !this.showNeighbors();
    if (next) {
      this.showGridTools.set(false);
      this.showAreaDetails.set(false);
      this.showTransports.set(false);
    } else {
      this.showGridTools.set(true);
    }
    this.showNeighbors.set(next);
  }

  protected toggleGridTools(): void {
    const next = !this.showGridTools();
    if (next) {
      this.showAreaDetails.set(false);
      this.showTransports.set(false);
      this.showNeighbors.set(false);
    }
    this.showGridTools.set(next);
  }

  protected addTransport(): void {
    const area = this.currentArea();
    if (!area) {
      return;
    }
    area.Transports.push(this.createEmptyTransport());
    this.touchBootstrap();
  }

  protected duplicateTransport(index: number): void {
    const area = this.currentArea();
    if (!area || index < 0 || index >= area.Transports.length) {
      return;
    }
    const existing = area.Transports[index];
    area.Transports.push({ ...existing });
    this.touchBootstrap();
  }

  protected deleteTransport(index: number): void {
    const area = this.currentArea();
    if (!area || index < 0 || index >= area.Transports.length) {
      return;
    }
    area.Transports.splice(index, 1);
    this.touchBootstrap();
  }

  protected toggleNavigationMap(): void {
    this.showNavigationMap.update((value) => !value);
  }

  protected startSideColumnResize(event: PointerEvent): void {
    event.preventDefault();
    const startX = event.clientX;
    const startWidth = this.sideColumnWidth();
    this.isResizingSideColumn.set(true);

    const onPointerMove = (moveEvent: PointerEvent) => {
      const delta = startX - moveEvent.clientX;
      const nextWidth = Math.max(280, Math.min(760, startWidth + delta));
      this.sideColumnWidth.set(nextWidth);
    };

    const onPointerUp = () => {
      this.isResizingSideColumn.set(false);
      window.removeEventListener('pointermove', onPointerMove);
      window.removeEventListener('pointerup', onPointerUp);
    };

    window.addEventListener('pointermove', onPointerMove);
    window.addEventListener('pointerup', onPointerUp);
  }

  protected hasNavigationImageError(areaName: string): boolean {
    const key = this.navigationImageKey(areaName);
    return Boolean(this.failedNavigationImageKeys()[key]);
  }

  protected onNavigationImageError(areaName: string): void {
    const key = this.navigationImageKey(areaName);
    this.failedNavigationImageKeys.update((entries) => ({ ...entries, [key]: true }));
  }

  protected resetAreaEditPanels(): void {
    this.showGridTools.set(true);
    this.showAreaDetails.set(false);
    this.showTransports.set(false);
    this.showNeighbors.set(false);
  }

  protected onCollectionChange(): void {
    this.spaceName.set(this.spaceNames()[0] ?? '');
    this.modifySpaceName.set(this.spaceNames()[0] ?? '');
    this.onSpaceChange();

    this.prototypeSet.set(this.prototypeSets()[0] ?? '');
    this.fragmentSet.set(this.fragmentSets()[0] ?? '');
    this.interactableSet.set(this.interactableSets()[0] ?? '');

    this.onPrototypeSetChange();
    this.onFragmentSetChange();
    this.onInteractableSetChange();
  }

  protected onSpaceChange(): void {
    this.resetAreaEditPanels();
    this.areaName.set(this.areaNames()[0] ?? '');
    this.failedNavigationImageKeys.set({});
    const area = this.currentArea();
    if (area) {
      this.newArea.set({
        ...this.newArea(),
        defaultTileColor: area.Blueprint.DefaultTileColor || 'green',
        defaultTileColor1: area.Blueprint.DefaultTileColor1 || 'brown',
      });
    }
    this.selection.set(undefined);
    this.hoverPosition.set(undefined);
  }

  protected onAreaChange(): void {
    this.selection.set(undefined);
    this.hoverPosition.set(undefined);
  }

  protected jumpToTransportDestination(destinationAreaName: string | undefined): void {
    this.goToArea(destinationAreaName);
  }

  protected goToArea(targetAreaName: string | undefined, event?: Event): void {
    event?.preventDefault();
    const candidate = (targetAreaName ?? '').trim();
    if (!candidate) {
      return;
    }
    const destination = this.findAreaDestination(candidate);
    if (!destination) {
      return;
    }
    if (this.spaceName() !== destination.spaceName) {
      this.spaceName.set(destination.spaceName);
      this.failedNavigationImageKeys.set({});
    }
    this.areaName.set(destination.areaName);
    this.onAreaChange();
  }

  protected onFixtureChange(): void {
    this.selection.set(undefined);
    this.hoverPosition.set(undefined);
    switch (this.fixture()) {
      case 'prototype':
        this.tool.set('select');
        this.selectedAssetId.set(this.prototypes()[0]?.id ?? '');
        break;
      case 'fragment':
        this.tool.set('place-blueprint');
        this.selectedAssetId.set(this.fragments()[0]?.id ?? '');
        break;
      case 'interactable':
        this.tool.set('interactable-replace');
        this.selectedAssetId.set(this.interactables()[0]?.id ?? '');
        this.onSelectedInteractableAssetChange();
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
    this.ensureInteractableStateSelection();
    this.ensureEditedInteractableStateSelection();
  }

  protected onSelectedInteractableAssetChange(): void {
    this.ensureInteractableStateSelection();
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
      selectedInteractableState: this.selectedInteractableState(),
      selected: this.selection(),
      blueprint,
      prototypesById: this.prototypesById(),
      fragmentsById: this.fragmentsById(),
      interactablesById: this.interactablesById(),
    });

    this.selection.set(nextSelection);
    if (nextSelection) {
      this.showSelectedInformation.set(true);
    }
    this.touchBootstrap();
  }

  protected onGridHover(y: number, x: number): void {
    this.hoverPosition.set({ y, x });
  }

  protected onGridLeave(): void {
    this.hoverPosition.set(undefined);
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
    const index = blueprint.Instructions.length - 1;
    updateInstructionAndReapply(blueprint, index, {}, this.prototypesById(), this.fragmentsById());
    this.markInstructionEdited(blueprint.Instructions[index].ID);
    this.touchBootstrap();
  }

  protected moveInstruction(index: number): void {
    const blueprint = this.activeBlueprint();
    if (!blueprint || !blueprint.Instructions || blueprint.Instructions.length < 2) {
      return;
    }
    reorderInstructionAndReapply(blueprint, index, this.prototypesById(), this.fragmentsById());
    this.touchBootstrap();
  }

  protected deleteInstruction(index: number): void {
    const blueprint = this.activeBlueprint();
    if (!blueprint) {
      return;
    }
    const deletedId = blueprint.Instructions[index]?.ID;
    deleteInstructionAndReapply(blueprint, index, this.prototypesById(), this.fragmentsById());
    if (deletedId) {
      this.instructionEditedIds.update((current) => {
        const next = { ...current };
        delete next[deletedId];
        return next;
      });
    }
    this.touchBootstrap();
  }

  protected updateInstructionY(index: number, value: number | string): void {
    const blueprint = this.activeBlueprint();
    const instruction = blueprint?.Instructions[index];
    if (!blueprint || !instruction) {
      return;
    }
    updateInstructionAndReapply(
      blueprint,
      index,
      { Y: normalizeInstructionField(value, instruction.Y) },
      this.prototypesById(),
      this.fragmentsById(),
    );
    this.markInstructionEdited(instruction.ID);
    this.touchBootstrap();
  }

  protected updateInstructionX(index: number, value: number | string): void {
    const blueprint = this.activeBlueprint();
    const instruction = blueprint?.Instructions[index];
    if (!blueprint || !instruction) {
      return;
    }
    updateInstructionAndReapply(
      blueprint,
      index,
      { X: normalizeInstructionField(value, instruction.X) },
      this.prototypesById(),
      this.fragmentsById(),
    );
    this.markInstructionEdited(instruction.ID);
    this.touchBootstrap();
  }

  protected updateInstructionRotation(index: number, value: number | string): void {
    const blueprint = this.activeBlueprint();
    const instruction = blueprint?.Instructions[index];
    if (!blueprint || !instruction) {
      return;
    }
    updateInstructionAndReapply(
      blueprint,
      index,
      { ClockwiseRotations: normalizeInstructionField(value, instruction.ClockwiseRotations) },
      this.prototypesById(),
      this.fragmentsById(),
    );
    this.markInstructionEdited(instruction.ID);
    this.touchBootstrap();
  }

  protected instructionAssetLabel(assetId: string): string {
    const resolved = this.resolveAssetById(assetId);
    if (!resolved) {
      return `Unknown asset (${assetId || 'empty'})`;
    }
    return this.assetDisplayName(resolved);
  }

  protected instructionHasKnownAsset(assetId: string): boolean {
    return Boolean(this.resolveAssetById(assetId));
  }

  protected instructionIsEdited(instructionId: string): boolean {
    return Boolean(this.instructionEditedIds()[instructionId]);
  }

  protected showInstruction(index: number): void {
    const blueprint = this.activeBlueprint();
    if (!blueprint?.Instructions[index]) {
      return;
    }
    const instruction = blueprint.Instructions[index];
    this.selection.set({ y: instruction.Y, x: instruction.X });
  }

  protected toggleBlueprintInstructions(): void {
    this.showBlueprintInstructions.update((value) => !value);
  }

  protected toggleSelectedInformation(): void {
    this.showSelectedInformation.update((value) => !value);
  }

  protected openPrototypeDetails(prototypeId: string, event?: Event): void {
    event?.preventDefault();
    const prototype = this.prototypesById().get(prototypeId);
    if (!prototype) {
      return;
    }
    this.prototypeSet.set(prototype.setName);
    this.onPrototypeSetChange();
    this.prototypeEditId.set(prototype.id);
    this.selectedAssetId.set(prototype.id);
    this.setViewMode('prototypes');
  }

  protected openInteractableDetails(interactableId: string, event?: Event): void {
    event?.preventDefault();
    const interactable = this.interactablesById().get(interactableId);
    if (!interactable) {
      return;
    }
    this.interactableSet.set(interactable.setName);
    this.onInteractableSetChange();
    this.interactableEditId.set(interactable.id);
    this.selectedAssetId.set(interactable.id);
    this.ensureEditedInteractableStateSelection();
    this.setViewMode('interactables');
  }

  protected openInstructionAssetDetails(assetId: string, event?: Event): void {
    event?.preventDefault();
    const resolved = this.resolveAssetById(assetId);
    if (!resolved) {
      return;
    }

    if (resolved.type === 'prototype') {
      this.openPrototypeDetails(resolved.asset.id);
      return;
    }

    if (resolved.type === 'fragment') {
      this.fragmentSet.set(resolved.asset.setName);
      this.onFragmentSetChange();
      this.fragmentEditId.set(resolved.asset.id);
      this.selectedAssetId.set(resolved.asset.id);
      this.setViewMode('fragments');
      return;
    }

    this.openInteractableDetails(resolved.asset.id);
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

  protected async applyAreaPropertyToSpace(): Promise<void> {
    const colName = this.collectionName();
    const sName = this.modifySpaceName();
    const space = this.modifyTargetSpace();
    if (!colName || !sName || !space) {
      return;
    }

    const property = this.bulkAreaProperty();
    const textValue = this.bulkAreaValueText();
    const boolValue = this.bulkAreaValueBoolean();

    for (const area of space.Areas) {
      switch (property) {
        case 'safe':
          area.Safe = boolValue;
          break;
        case 'defaultTileColor':
          area.Blueprint.DefaultTileColor = textValue;
          break;
        case 'defaultTileColor1':
          area.Blueprint.DefaultTileColor1 = textValue;
          break;
        case 'weather':
          area.Weather = textValue;
          break;
        case 'loadStrategy':
          area.LoadStrategy = textValue;
          break;
        case 'spawnStrategy':
          area.SpawnStrategy = textValue;
          break;
        case 'broadcastGroup':
          area.BroadcastGroup = textValue;
          break;
      }
    }

    this.touchBootstrap();
    this.status.set('Saving space...');
    await this.api.saveSpace(colName, sName, space);
    this.status.set(`Updated ${space.Areas.length} areas in ${sName}.`);
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
      state: 'default',
      defaultState: 'default',
      states: {
        default: {
          cssClass: '',
          pushable: false,
          walkable: false,
          fragile: false,
          rejectTeleport: false,
          reactions: '',
          reactionRules: [],
        },
      },
      cssClass: '',
      pushable: false,
      walkable: false,
      fragile: false,
      reactions: '',
      reactionRules: [],
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

  // ── Reaction rule helpers ────────────────────────────────────────────

  protected addReactionRule(): void {
    const state = this.editedInteractableState();
    if (!state) return;
    if (!state.reactionRules) {
      state.reactionRules = [];
    }
    state.reactionRules.push({
      reactsWith: 'everything',
      reactsWithArgs: [],
      reaction: 'pass',
      reactionArgs: [],
    });
    const interactable = this.editedInteractable();
    if (interactable) {
      this.syncTopLevelFromCurrentState(interactable);
    }
    this.touchBootstrap();
  }

  protected deleteReactionRule(index: number): void {
    const state = this.editedInteractableState();
    if (!state?.reactionRules) return;
    state.reactionRules.splice(index, 1);
    const interactable = this.editedInteractable();
    if (interactable) {
      this.syncTopLevelFromCurrentState(interactable);
    }
    this.touchBootstrap();
  }

  protected moveReactionRuleUp(index: number): void {
    const state = this.editedInteractableState();
    if (!state?.reactionRules || index <= 0) return;
    const rules = state.reactionRules;
    [rules[index - 1], rules[index]] = [rules[index], rules[index - 1]];
    const interactable = this.editedInteractable();
    if (interactable) {
      this.syncTopLevelFromCurrentState(interactable);
    }
    this.touchBootstrap();
  }

  protected onReactsWithChange(rule: ReactionRule, newKey: string): void {
    rule.reactsWith = newKey;
    const entry = findReactsWithEntry(newKey);
    rule.reactsWithArgs = entry ? entry.args.map(() => '') : [];
    const interactable = this.editedInteractable();
    if (interactable) {
      this.syncTopLevelFromCurrentState(interactable);
    }
    this.touchBootstrap();
  }

  protected onReactionChange(rule: ReactionRule, newKey: string): void {
    rule.reaction = newKey;
    const entry = findReactionEntry(newKey);
    rule.reactionArgs = entry ? entry.args.map(() => '') : [];
    const interactable = this.editedInteractable();
    if (interactable) {
      this.syncTopLevelFromCurrentState(interactable);
    }
    this.touchBootstrap();
  }

  protected reactsWithArgsFor(key: string): RegistryEntry['args'] {
    return findReactsWithEntry(key)?.args ?? [];
  }

  protected reactionArgsFor(key: string): RegistryEntry['args'] {
    return findReactionEntry(key)?.args ?? [];
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

  protected async flattenSpace(): Promise<void> {
    const colName = this.collectionName();
    const sName = this.spaceName();
    const space = this.currentSpace();
    if (!colName || !sName || !space) {
      return;
    }
    if (!this.isSimplyTiledSpace(space)) {
      this.status.set('Only simply tiled spaces may be flattened.');
      return;
    }

    this.status.set('Flattening space...');
    const result = await this.api.flattenSpace(colName, sName);
    const flattenedName = result.spaceName;
    await this.loadBootstrap();
    this.spaceName.set(flattenedName);
    this.modifySpaceName.set(flattenedName);
    this.onSpaceChange();
    this.status.set(`Space flattened into ${flattenedName}.`);
  }

  protected async resetUnsavedChanges(): Promise<void> {
    const previousCollectionName = this.collectionName();
    const previousSpaceName = this.spaceName();
    const previousAreaName = this.areaName();
    if (!previousCollectionName || !previousSpaceName) {
      return;
    }

    this.status.set('Resetting unsaved space changes...');
    await this.loadBootstrap();

    if (this.collectionNames().includes(previousCollectionName)) {
      this.collectionName.set(previousCollectionName);
      this.onCollectionChange();
    }

    if (this.spaceNames().includes(previousSpaceName)) {
      this.spaceName.set(previousSpaceName);
      this.modifySpaceName.set(previousSpaceName);
      this.onSpaceChange();
    }

    if (this.areaNames().includes(previousAreaName)) {
      this.areaName.set(previousAreaName);
      this.onAreaChange();
    }

    this.instructionEditedIds.set({});

    this.status.set(`Unsaved changes in ${previousSpaceName} reset.`);
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

  protected ensureEditedInteractableStateSelection(): void {
    const interactable = this.editedInteractable();
    if (!interactable) {
      this.interactableStateEditName.set('default');
      return;
    }
    this.ensureInteractableStateModel(interactable);
    const candidate = this.interactableStateEditName() || interactable.defaultState || 'default';
    if (interactable.states?.[candidate]) {
      this.interactableStateEditName.set(candidate);
      return;
    }
    this.interactableStateEditName.set(interactable.defaultState || 'default');
  }

  protected addEditedInteractableState(): void {
    const interactable = this.editedInteractable();
    if (!interactable) {
      return;
    }
    this.ensureInteractableStateModel(interactable);

    const baseName = 'state';
    let index = 1;
    let nextName = `${baseName}-${index}`;
    while (interactable.states?.[nextName]) {
      index += 1;
      nextName = `${baseName}-${index}`;
    }

    interactable.states![nextName] = {
      ...(interactable.states?.[interactable.defaultState || 'default'] ?? {
        cssClass: '',
        pushable: false,
        walkable: false,
        fragile: false,
        rejectTeleport: false,
        reactions: '',
        reactionRules: [],
      }),
    };

    this.interactableStateEditName.set(nextName);
    this.syncTopLevelFromCurrentState(interactable);
    this.touchBootstrap();
  }

  protected deleteEditedInteractableState(): void {
    const interactable = this.editedInteractable();
    if (!interactable) {
      return;
    }
    this.ensureInteractableStateModel(interactable);

    const stateName = this.interactableStateEditName();
    const stateNames = Object.keys(interactable.states ?? {});
    if (stateNames.length <= 1 || !stateName || !interactable.states?.[stateName]) {
      return;
    }

    delete interactable.states[stateName];
    if (interactable.defaultState === stateName) {
      interactable.defaultState = Object.keys(interactable.states)[0] ?? 'default';
    }

    this.interactableStateEditName.set(interactable.defaultState || 'default');
    this.syncTopLevelFromCurrentState(interactable);
    this.touchBootstrap();
  }

  protected onEditedInteractableStateNameChange(): void {
    const interactable = this.editedInteractable();
    if (!interactable) {
      return;
    }
    this.ensureInteractableStateModel(interactable);
    this.syncTopLevelFromCurrentState(interactable);
    this.touchBootstrap();
  }

  protected onEditedInteractableDefaultStateChange(): void {
    const interactable = this.editedInteractable();
    if (!interactable) {
      return;
    }
    this.ensureInteractableStateModel(interactable);
    if (!interactable.states?.[interactable.defaultState || '']) {
      interactable.defaultState = Object.keys(interactable.states ?? {})[0] ?? 'default';
    }
    this.syncTopLevelFromCurrentState(interactable);
    this.touchBootstrap();
  }

  protected setSelectedTileInteractableState(stateName: string): void {
    const tile = this.selectedTile();
    const interactable = this.selectedTileInteractable();
    if (!tile || !interactable || !interactable.states?.[stateName]) {
      return;
    }
    tile.interactableState = stateName;
    this.touchBootstrap();
  }

  protected selectedTileInteractableStateNames(): string[] {
    return Object.keys(this.selectedTileInteractable()?.states ?? {});
  }

  protected selectedInteractableAssetStateNames(): string[] {
    const interactable = this.interactablesById().get(this.selectedAssetId());
    if (!interactable) {
      return ['default'];
    }
    this.ensureInteractableStateModel(interactable);
    return Object.keys(
      interactable.states ?? {
        default: {
          cssClass: '',
          pushable: false,
          walkable: false,
          fragile: false,
          rejectTeleport: false,
          reactions: '',
          reactionRules: [],
        },
      },
    );
  }

  private ensureInteractableStateSelection(): void {
    const interactable = this.interactablesById().get(this.selectedAssetId());
    if (!interactable) {
      this.selectedInteractableState.set('default');
      return;
    }

    this.ensureInteractableStateModel(interactable);
    const candidate = this.selectedInteractableState() || interactable.defaultState || 'default';
    if (interactable.states?.[candidate]) {
      this.selectedInteractableState.set(candidate);
      return;
    }
    this.selectedInteractableState.set(interactable.defaultState || 'default');
  }

  private ensureInteractableStateModel(interactable: InteractableDescription): void {
    const defaultState = (interactable.defaultState || 'default').trim() || 'default';
    interactable.defaultState = defaultState;
    if (!interactable.states) {
      interactable.states = {};
    }
    if (!interactable.states[defaultState]) {
      interactable.states[defaultState] = {
        cssClass: interactable.cssClass ?? '',
        pushable: Boolean(interactable.pushable),
        walkable: Boolean(interactable.walkable),
        fragile: Boolean(interactable.fragile),
        rejectTeleport: Boolean(interactable.rejectTeleport),
        reactions: interactable.reactions ?? '',
        reactionRules: interactable.reactionRules ?? [],
      };
    }

    if (!interactable.state || !interactable.states[interactable.state]) {
      interactable.state = defaultState;
    }

    this.syncTopLevelFromCurrentState(interactable);
  }

  private syncTopLevelFromCurrentState(interactable: InteractableDescription): void {
    this.ensureStateRuleArrays(interactable);
    const stateName = this.interactableStateEditName() || interactable.state || interactable.defaultState || 'default';
    const config = interactable.states?.[stateName];
    if (!config) {
      return;
    }

    interactable.state = stateName;
    interactable.cssClass = config.cssClass;
    interactable.pushable = config.pushable;
    interactable.walkable = config.walkable;
    interactable.fragile = config.fragile;
    interactable.rejectTeleport = config.rejectTeleport;
    interactable.reactions = config.reactions;
    interactable.reactionRules = config.reactionRules;
  }

  private ensureStateRuleArrays(interactable: InteractableDescription): void {
    if (!interactable.states) {
      return;
    }
    for (const stateName of Object.keys(interactable.states)) {
      const state = interactable.states[stateName];
      if (!state.reactionRules) {
        state.reactionRules = [];
      }
    }
  }

  private resolveInteractableForAssetAndState(interactableId: string, tileStateName?: string): InteractableDescription | undefined {
    const base = this.interactablesById().get((interactableId ?? '').trim());
    if (!base) {
      return undefined;
    }

    this.ensureInteractableStateModel(base);
    const defaultState = base.defaultState || 'default';
    const selectedState = (tileStateName || base.state || defaultState).trim();
    const config = base.states?.[selectedState] ?? base.states?.[defaultState];
    if (!config) {
      return base;
    }

    return {
      ...base,
      state: selectedState && base.states?.[selectedState] ? selectedState : defaultState,
      cssClass: config.cssClass,
      pushable: config.pushable,
      walkable: config.walkable,
      fragile: config.fragile,
      rejectTeleport: config.rejectTeleport,
      reactions: config.reactions,
      reactionRules: config.reactionRules,
    };
  }

  private resolveInteractableForTile(tile: { interactableId?: string; interactableState?: string }): InteractableDescription | undefined {
    const interactableId = (tile.interactableId ?? '').trim();
    if (!interactableId) {
      return undefined;
    }
    return this.resolveInteractableForAssetAndState(interactableId, tile.interactableState ?? '');
  }

  private clampColorChannel(value: unknown): number {
    const parsed = Number(value);
    if (Number.isNaN(parsed)) {
      return 0;
    }
    return Math.max(0, Math.min(255, Math.round(parsed)));
  }

  private clampAlpha(value: unknown): number {
    if (value === null || value === undefined || value === '') {
      return 1;
    }
    const parsed = Number(value);
    if (Number.isNaN(parsed)) {
      return 1;
    }
    return Math.max(0, Math.min(1, parsed));
  }

  private isSimplyTiledSpace(space: Space): boolean {
    return space.Topology === 'plane' || space.Topology === 'torus';
  }

  private buildAreaImageUrl(spaceName: string, areaName: string, collectionName: string): string {
    return `/images/make/${encodeURIComponent(spaceName)}/${encodeURIComponent(areaName)}?currentCollection=${encodeURIComponent(collectionName)}`;
  }

  private navigationImageKey(areaName: string): string {
    return `${this.collectionName()}::${this.spaceName()}::${areaName}`;
  }

  private resolveAssetById(
    assetId: string,
  ):
    | { type: 'prototype'; asset: Prototype }
    | { type: 'fragment'; asset: Fragment }
    | { type: 'interactable'; asset: InteractableDescription }
    | undefined {
    const trimmed = assetId.trim();
    if (!trimmed) {
      return undefined;
    }

    const prototype = this.prototypesById().get(trimmed);
    if (prototype) {
      return { type: 'prototype', asset: prototype };
    }

    const fragment = this.fragmentsById().get(trimmed);
    if (fragment) {
      return { type: 'fragment', asset: fragment };
    }

    const interactable = this.interactablesById().get(trimmed);
    if (interactable) {
      return { type: 'interactable', asset: interactable };
    }

    return undefined;
  }

  private assetDisplayName(
    resolved:
      | { type: 'prototype'; asset: Prototype }
      | { type: 'fragment'; asset: Fragment }
      | { type: 'interactable'; asset: InteractableDescription },
  ): string {
    if (resolved.type === 'prototype') {
      return `Prototype: ${resolved.asset.commonName || resolved.asset.id}`;
    }
    if (resolved.type === 'fragment') {
      return `Fragment: ${resolved.asset.name || resolved.asset.id}`;
    }
    return `Interactable: ${resolved.asset.name || resolved.asset.id}`;
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

  private markInstructionEdited(instructionId: string): void {
    if (!instructionId) {
      return;
    }
    this.instructionEditedIds.update((current) => ({ ...current, [instructionId]: true }));
  }

  private async loadBootstrap(): Promise<void> {
    this.loading.set(true);
    const raw = await this.api.getBootstrap();
    const data = this.normalizeBootstrap(raw as unknown as Record<string, unknown>);
    this.bootstrap.set(data);

    this.collectionName.set(this.collectionNames()[0] ?? '');
    this.spaceName.set(this.spaceNames()[0] ?? '');
    this.modifySpaceName.set(this.spaceNames()[0] ?? '');
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
    const ground = input?.Ground ?? input?.ground;
    return {
      Tiles: input?.Tiles ?? input?.tiles ?? [],
      Instructions: input?.Instructions ?? input?.instructions ?? [],
      Ground: Array.isArray(ground)
        ? ground.map((row: any[]) =>
            row.map((cell: any): Cell => ({
              status: Number(cell?.status ?? cell?.Status ?? 0),
              topLeft: Boolean(cell?.topLeft ?? cell?.TopLeft),
              topRight: Boolean(cell?.topRight ?? cell?.TopRight),
              bottomLeft: Boolean(cell?.bottomLeft ?? cell?.BottomLeft),
              bottomRight: Boolean(cell?.bottomRight ?? cell?.BottomRight),
            })),
          )
        : undefined,
      DefaultTileColor: input?.DefaultTileColor ?? input?.defaultTileColor ?? 'white',
      DefaultTileColor1: input?.DefaultTileColor1 ?? input?.defaultTileColor1 ?? 'white',
    };
  }

  private findAreaDestination(areaName: string): { spaceName: string; areaName: string } | undefined {
    const collection = this.currentCollection();
    if (!collection) {
      return undefined;
    }

    const currentSpaceName = this.spaceName();
    const currentSpace = collection.Spaces[currentSpaceName];
    if (currentSpace?.Areas.some((area) => area.Name === areaName)) {
      return { spaceName: currentSpaceName, areaName };
    }

    for (const spaceName of Object.keys(collection.Spaces)) {
      if (spaceName === currentSpaceName) {
        continue;
      }
      const space = collection.Spaces[spaceName];
      if (space?.Areas.some((area) => area.Name === areaName)) {
        return { spaceName, areaName };
      }
    }

    return undefined;
  }

  private normalizeTransport(input: any): Transport {
    return {
      SourceY: Number(input?.SourceY ?? input?.sourceY ?? 0),
      SourceX: Number(input?.SourceX ?? input?.sourceX ?? 0),
      DestY: Number(input?.DestY ?? input?.destY ?? 0),
      DestX: Number(input?.DestX ?? input?.destX ?? 0),
      DestStage: input?.DestStage ?? input?.destStage ?? '',
      Confirmation: Boolean(input?.Confirmation ?? input?.confirmation),
      RejectInteractable: Boolean(input?.RejectInteractable ?? input?.rejectInteractable),
    };
  }

  private createEmptyTransport(): Transport {
    return {
      SourceY: 0,
      SourceX: 0,
      DestY: 0,
      DestX: 0,
      DestStage: '',
      Confirmation: false,
      RejectInteractable: false,
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
            Transports: (area.Transports ?? area.transports ?? []).map((transport: any) => this.normalizeTransport(transport)),
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

      const outInteractableSets: Record<string, InteractableDescription[]> = {};
      const sourceInteractableSets = (collection.InteractableSets ?? collection.interactableSets ?? {}) as Record<string, any[]>;
      for (const setName of Object.keys(sourceInteractableSets)) {
        outInteractableSets[setName] = (sourceInteractableSets[setName] ?? []).map((entry: any) => {
          const legacyRules = (entry.reactionRules ?? entry.ReactionRules ?? []).map((rule: any) => ({
            reactsWith: rule.reactsWith ?? '',
            reactsWithArgs: rule.reactsWithArgs ?? [],
            reaction: rule.reaction ?? '',
            reactionArgs: rule.reactionArgs ?? [],
          }));

          const normalizedStates: Record<string, InteractableStateDescription> = {};
          const sourceStates = entry.states ?? entry.States ?? {};
          for (const stateName of Object.keys(sourceStates)) {
            const state = sourceStates[stateName] ?? {};
            normalizedStates[stateName] = {
              cssClass: state.cssClass ?? state.CssClass ?? '',
              pushable: Boolean(state.pushable ?? state.Pushable),
              walkable: Boolean(state.walkable ?? state.Walkable),
              fragile: Boolean(state.fragile ?? state.Fragile),
              rejectTeleport: Boolean(state.rejectTeleport ?? state.RejectTeleport),
              reactions: state.reactions ?? state.Reactions ?? '',
              reactionRules: (state.reactionRules ?? state.ReactionRules ?? []).map((rule: any) => ({
                reactsWith: rule.reactsWith ?? '',
                reactsWithArgs: rule.reactsWithArgs ?? [],
                reaction: rule.reaction ?? '',
                reactionArgs: rule.reactionArgs ?? [],
              })),
            };
          }

          const defaultState = (entry.defaultState ?? entry.DefaultState ?? 'default') as string;
          if (!normalizedStates[defaultState]) {
            normalizedStates[defaultState] = {
              cssClass: entry.cssClass ?? entry.CssClass ?? '',
              pushable: Boolean(entry.pushable ?? entry.Pushable),
              walkable: Boolean(entry.walkable ?? entry.Walkable),
              fragile: Boolean(entry.fragile ?? entry.Fragile),
              rejectTeleport: Boolean(entry.rejectTeleport ?? entry.RejectTeleport),
              reactions: entry.reactions ?? entry.Reactions ?? '',
              reactionRules: legacyRules,
            };
          }

          const selectedState = (entry.state ?? entry.State ?? defaultState) as string;
          const selectedConfig = normalizedStates[selectedState] ?? normalizedStates[defaultState];

          return {
            id: entry.id ?? entry.ID ?? crypto.randomUUID(),
            name: entry.name ?? entry.Name ?? '',
            setName: entry.setName ?? entry.SetName ?? setName,
            state: selectedState,
            defaultState,
            states: normalizedStates,
            cssClass: selectedConfig.cssClass,
            pushable: selectedConfig.pushable,
            walkable: selectedConfig.walkable,
            fragile: selectedConfig.fragile,
            rejectTeleport: selectedConfig.rejectTeleport,
            reactions: selectedConfig.reactions,
            reactionRules: selectedConfig.reactionRules,
          };
        });
      }

      outCollections[collectionName] = {
        Name: collection.Name ?? collection.name ?? collectionName,
        Spaces: spaces,
        Fragments: outFragments,
        PrototypeSets: (collection.PrototypeSets ?? collection.prototypeSets ?? {}) as Collection['PrototypeSets'],
        InteractableSets: outInteractableSets,
      };
    }

    return {
      collections: outCollections,
      colors: (raw['colors'] ?? []) as BootstrapResponse['colors'],
    };
  }
}
