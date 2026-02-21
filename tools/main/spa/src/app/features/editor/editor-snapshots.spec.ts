import { ComponentFixture, TestBed } from '@angular/core/testing';
import { EditorComponent } from './editor.component';
import { EditorApiService } from '../../core/services/editor-api.service';
import {
  AreaDescription,
  BootstrapResponse,
  InteractableDescription,
  Material,
  Prototype,
  TileData,
} from '../../core/models/editor.models';
import { Tool, generateMaterials } from './grid-engine';

describe('EditorComponent snapshots', () => {
  let component: EditorComponent;
  let fixture: ComponentFixture<EditorComponent>;
  let api: jest.Mocked<EditorApiService>;

  beforeEach(async () => {
    api = {
      createCollection: jest.fn().mockResolvedValue(undefined),
      createSpace: jest.fn().mockResolvedValue(undefined),
      createArea: jest.fn().mockResolvedValue(undefined),
      getBootstrap: jest.fn().mockImplementation(async () => buildBootstrap()),
      saveSpace: jest.fn().mockResolvedValue(undefined),
      flattenSpace: jest.fn().mockResolvedValue({ spaceName: 'snap-space-flat' }),
      savePrototypeSet: jest.fn().mockResolvedValue(undefined),
      saveFragmentSet: jest.fn().mockResolvedValue(undefined),
      saveInteractableSet: jest.fn().mockResolvedValue(undefined),
      saveColors: jest.fn().mockResolvedValue(undefined),
      compile: jest.fn().mockResolvedValue(undefined),
      deploy: jest.fn().mockResolvedValue(undefined),
    } as unknown as jest.Mocked<EditorApiService>;

    await TestBed.configureTestingModule({
      imports: [EditorComponent],
      providers: [{ provide: EditorApiService, useValue: api }],
    }).compileComponents();

    fixture = TestBed.createComponent(EditorComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
    fixture.detectChanges();
  });

  it('captures baseline area snapshots', () => {
    const state = component as any;

    setArea(state, 'random-room');
    expect(serializeArea(state)).toMatchSnapshot('baseline random-room');

    setArea(state, 'staged-room');
    expect(serializeArea(state)).toMatchSnapshot('baseline staged-room');
  });

  it('captures snapshot after a sequence of editor UI actions', () => {
    const state = component as any;
    const randomSpy = jest
      .spyOn(crypto, 'randomUUID')
      .mockReturnValueOnce('00000000-0000-4000-8000-000000000001')
      .mockReturnValueOnce('00000000-0000-4000-8000-000000000002');

    setArea(state, 'random-room');

    clickGrid(state, { y: 0, x: 0, fixture: 'prototype', tool: 'fill', selectedAssetId: 'p-sand' });
    clickGrid(state, { y: 6, x: 1, fixture: 'prototype', tool: 'between', selectedAssetId: 'p-wall', selected: { y: 1, x: 7 } });
    clickGrid(state, { y: 2, x: 2, fixture: 'prototype', tool: 'between', selectedAssetId: 'p-grass', selected: { y: 5, x: 7 } });

    clickGrid(state, { y: 4, x: 6, fixture: 'prototype', tool: 'between', selectedAssetId: 'p-wall', selected: { y: 1, x: 6 } });
    clickGrid(state, { y: 4, x: 6, fixture: 'prototype', tool: 'between', selectedAssetId: 'p-wall', selected: { y: 4, x: 3 } });
    clickGrid(state, { y: 2, x: 4, fixture: 'prototype', tool: 'between', selectedAssetId: 'p-wall', selected: { y: 4, x: 4 } });

    clickGrid(state, { y: 2, x: 7, fixture: 'prototype', tool: 'fill', selectedAssetId: '' });

    clickGrid(state, { y: 6, x: 6, fixture: 'ground', tool: 'toggle-fill', selectedAssetId: '' });
    clickGrid(state, { y: 2, x: 3, fixture: 'ground', tool: 'toggle-between', selectedAssetId: '', selected: { y: 1, x: 1 } });
    clickGrid(state, { y: 5, x: 6, fixture: 'ground', tool: 'toggle', selectedAssetId: '' });

    clickGrid(state, { y: 6, x: 0, fixture: 'fragment', tool: 'place', selectedAssetId: 'frag-2x2' });
    clickGrid(state, { y: 6, x: 2, fixture: 'fragment', tool: 'place-blueprint', selectedAssetId: 'frag-2x2' });
    clickGrid(state, { y: 5, x: 2, fixture: 'prototype', tool: 'place-blueprint', selectedAssetId: 'p-blue' });

    clickGrid(state, { y: 5, x: 2, fixture: 'transformation', tool: 'rotate', selectedAssetId: '' });
    clickGrid(state, { y: 6, x: 3, fixture: 'interactable', tool: 'interactable-replace', selectedAssetId: 'i-powder' });
    clickGrid(state, { y: 6, x: 3, fixture: 'interactable', tool: 'interactable-delete', selectedAssetId: '' });

    expect(serializeArea(state)).toMatchSnapshot('random-room after UI action sequence');
    randomSpy.mockRestore();
  });

  it('captures ghost highlight overlays for hover previews', () => {
    const state = component as any;

    setArea(state, 'random-room');

    state.fixture.set('prototype');
    state.tool.set('replace');
    state.selectedAssetId.set('p-round');
    state.onGridHover(2, 2);
    expect(serializeGhostState(state)).toMatchSnapshot('ghost prototype hover');

    state.fixture.set('fragment');
    state.tool.set('place');
    state.selectedAssetId.set('frag-2x2');
    state.onGridHover(5, 5);
    expect(serializeGhostState(state)).toMatchSnapshot('ghost fragment hover');

    state.fixture.set('interactable');
    state.tool.set('interactable-replace');
    state.selectedAssetId.set('i-powder');
    state.onGridHover(3, 4);
    expect(serializeGhostState(state)).toMatchSnapshot('ghost interactable hover');

    state.fixture.set('prototype');
    state.tool.set('select');
    state.selectedAssetId.set('p-round');
    state.onGridHover(1, 6);
    expect(serializeGhostState(state)).toMatchSnapshot('ghost select hover-marker only');
  });

  it('captures rounded-corner prototype rendering across rotations', () => {
    const state = component as any;

    setArea(state, 'random-room');

    clickGrid(state, { y: 4, x: 4, fixture: 'prototype', tool: 'replace', selectedAssetId: 'p-round' });
    expect(serializeRotationProbe(state, 4, 4)).toMatchSnapshot('rounded corners rotation 0');

    clickGrid(state, { y: 4, x: 4, fixture: 'transformation', tool: 'rotate', selectedAssetId: '' });
    expect(serializeRotationProbe(state, 4, 4)).toMatchSnapshot('rounded corners rotation 1');

    clickGrid(state, { y: 4, x: 4, fixture: 'transformation', tool: 'rotate', selectedAssetId: '' });
    expect(serializeRotationProbe(state, 4, 4)).toMatchSnapshot('rounded corners rotation 2');

    clickGrid(state, { y: 4, x: 4, fixture: 'transformation', tool: 'rotate', selectedAssetId: '' });
    expect(serializeRotationProbe(state, 4, 4)).toMatchSnapshot('rounded corners rotation 3');
  });
});

function clickGrid(
  state: any,
  input: {
    y: number;
    x: number;
    fixture: 'prototype' | 'fragment' | 'interactable' | 'transformation' | 'ground';
    tool: Tool;
    selectedAssetId: string;
    selected?: { y: number; x: number };
  },
): void {
  state.fixture.set(input.fixture);
  state.tool.set(input.tool);
  state.selectedAssetId.set(input.selectedAssetId);
  if (input.selected) {
    state.selection.set(input.selected);
  }
  state.onGridClick(input.y, input.x);
}

function setArea(state: any, areaName: string): void {
  state.areaName.set(areaName);
  state.onAreaChange();
}

function serializeArea(state: any): string {
  const area: AreaDescription | undefined = state.currentArea();
  if (!area) {
    return 'NO AREA';
  }

  const prototypesById: Map<string, Prototype> = state.prototypesById();
  const interactablesById: Map<string, InteractableDescription> = state.interactablesById();
  const materials = generateMaterials(area.Blueprint, prototypesById, false);

  const lines: string[] = [];
  lines.push(`Name: ${area.Name}`);
  lines.push(`Safe: ${area.Safe}`);
  lines.push(
    `North: ${area.North ?? ''} South: ${area.South ?? ''} East: ${area.East ?? ''} West: ${area.West ?? ''}`,
  );
  lines.push(`Weather: ${area.Weather ?? ''} BroadcastGroup: ${area.BroadcastGroup ?? ''}`);
  lines.push(`Transports: ${area.Transports.length}`);
  lines.push('');

  lines.push('Instructions:');
  if (!area.Blueprint.Instructions?.length) {
    lines.push('(none)');
  } else {
    for (let i = 0; i < area.Blueprint.Instructions.length; i += 1) {
      const instruction = area.Blueprint.Instructions[i];
      lines.push(
        `${i.toString().padStart(2, '0')}: (${instruction.Y},${instruction.X}) ${instruction.GridAssetId} r${instruction.ClockwiseRotations}`,
      );
    }
  }
  lines.push('');

  lines.push('Material Grid (walk:css:txt):');
  for (let y = 0; y < materials.length; y += 1) {
    lines.push(`r${y.toString().padStart(2, '0')} ${materials[y].map((material) => materialToken(material)).join(' | ')}`);
  }
  lines.push('');

  lines.push('Tile Grid (prototype/interactable/rotation):');
  for (let y = 0; y < area.Blueprint.Tiles.length; y += 1) {
    const row = area.Blueprint.Tiles[y];
    lines.push(
      `r${y.toString().padStart(2, '0')} ${row
        .map((tile) => {
          const proto = short(tile.prototypeId);
          const interactable = short(tile.interactableId);
          const rotation = tile.transformation?.clockwiseRotations ?? 0;
          return `${proto}/${interactable}/${rotation}`;
        })
        .join(' | ')}`,
    );
  }
  lines.push('');

  lines.push('Interactables (resolved):');
  for (let y = 0; y < area.Blueprint.Tiles.length; y += 1) {
    const row = area.Blueprint.Tiles[y];
    lines.push(
      `r${y.toString().padStart(2, '0')} ${row
        .map((tile) => {
          const id = tile.interactableId ?? '';
          if (!id) {
            return '-';
          }
          const found = interactablesById.get(id);
          return found?.name ?? id;
        })
        .join(' | ')}`,
    );
  }
  lines.push('');

  lines.push('Ground Grid:');
  if (!area.Blueprint.Ground?.length) {
    lines.push('(none)');
  } else {
    for (let y = 0; y < area.Blueprint.Ground.length; y += 1) {
      lines.push(
        `r${y.toString().padStart(2, '0')} ${area.Blueprint.Ground[y].map((cell) => groundToken(cell)).join(' | ')}`,
      );
    }
  }

  return lines.join('\n');
}

function serializeGhostState(state: any): string {
  const area: AreaDescription | undefined = state.currentArea();
  if (!area) {
    return 'NO AREA';
  }

  const ghostMaterials: (Material | undefined)[][] = state.ghostMaterials();
  const ghostInteractables: (InteractableDescription | undefined)[][] = state.ghostInteractables();
  const hover = state.hoverPosition();

  const lines: string[] = [];
  lines.push(`Fixture: ${state.fixture()} Tool: ${state.tool()} Asset: ${state.selectedAssetId() || '-'}`);
  lines.push(`Hover: ${hover ? `${hover.y},${hover.x}` : '-'}`);

  const overlays: string[] = [];
  for (let y = 0; y < area.Blueprint.Tiles.length; y += 1) {
    for (let x = 0; x < area.Blueprint.Tiles[y].length; x += 1) {
      const material = ghostMaterials[y]?.[x];
      const interactable = ghostInteractables[y]?.[x];
      if (!material && !interactable) {
        continue;
      }
      overlays.push(`(${y},${x}) m=${material ? materialToken(material) : '-'} i=${interactable ? interactable.name : '-'}`);
    }
  }

  lines.push('Overlays:');
  if (!overlays.length) {
    lines.push('(none)');
  } else {
    lines.push(...overlays);
  }

  const selectHover: string[] = [];
  for (let y = 0; y < area.Blueprint.Tiles.length; y += 1) {
    for (let x = 0; x < area.Blueprint.Tiles[y].length; x += 1) {
      if (state.shouldShowSelectHover(y, x)) {
        selectHover.push(`(${y},${x})`);
      }
    }
  }

  lines.push('SelectHover:');
  if (!selectHover.length) {
    lines.push('(none)');
  } else {
    lines.push(selectHover.join(' '));
  }

  return lines.join('\n');
}

function serializeRotationProbe(state: any, y: number, x: number): string {
  const area: AreaDescription | undefined = state.currentArea();
  if (!area) {
    return 'NO AREA';
  }

  const prototypesById: Map<string, Prototype> = state.prototypesById();
  const materials = generateMaterials(area.Blueprint, prototypesById, false);
  const tile = area.Blueprint.Tiles[y][x];
  const material = materials[y][x];

  const lines: string[] = [];
  lines.push(`Probe: (${y},${x})`);
  lines.push(`Tile: ${short(tile.prototypeId)}/${short(tile.interactableId)}@${tile.transformation?.clockwiseRotations ?? 0}`);
  lines.push(`MaterialToken: ${materialToken(material)}`);

  return lines.join('\n');
}

function materialToken(material: Material): string {
  const walkable = material.walkable ? 'T' : 'F';
  const css = short(chooseCss(material), 13);
  const hasText = material.displayText ? '1' : '0';
  return `${walkable}:${css}:${hasText}`;
}

function chooseCss(material: Material): string {
  return (
    material.ceiling2css ??
    material.ceiling1css ??
    material.layer2css ??
    material.layer1css ??
    material.ground2css ??
    material.ground1css ??
    ''
  );
}

function groundToken(cell: {
  status: number;
  topLeft?: boolean;
  topRight?: boolean;
  bottomLeft?: boolean;
  bottomRight?: boolean;
}): string {
  const flags: string[] = [];
  if (cell.topLeft) {
    flags.push('tl');
  }
  if (cell.topRight) {
    flags.push('tr');
  }
  if (cell.bottomLeft) {
    flags.push('bl');
  }
  if (cell.bottomRight) {
    flags.push('br');
  }
  if (!flags.length) {
    return `${cell.status}`;
  }
  return `${cell.status}[${flags.join(',')}]`;
}

function short(value: string | undefined, maxLength = 12): string {
  const clean = (value ?? '').trim();
  if (!clean) {
    return '-';
  }
  if (clean.length <= maxLength) {
    return clean;
  }
  return clean.slice(0, maxLength);
}

function makeTile(prototypeId: string, interactableId = ''): TileData {
  return {
    prototypeId,
    interactableId,
    transformation: { clockwiseRotations: 0 },
  };
}

function makeArea(name: string, size: number, prototypeId: string): AreaDescription {
  return {
    Name: name,
    Safe: true,
    Blueprint: {
      Tiles: Array.from({ length: size }, () => Array.from({ length: size }, () => makeTile(prototypeId))),
      Instructions: [],
      DefaultTileColor: 'white',
      DefaultTileColor1: 'sand',
    },
    Transports: [],
    North: '',
    South: '',
    East: '',
    West: '',
    Weather: '',
    LoadStrategy: '',
    SpawnStrategy: '',
    BroadcastGroup: '',
  };
}

function makePrototype(
  id: string,
  options: {
    commonName: string;
    walkable: boolean;
    cssColor: string;
    ceiling2css: string;
    displayText?: string;
    editorColor?: string;
  },
): Prototype {
  return {
    id,
    commonName: options.commonName,
    cssColor: options.cssColor,
    walkable: options.walkable,
    layer1css: '',
    layer2css: '',
    ceiling1css: '',
    ceiling2css: options.ceiling2css,
    setName: 'base-prototypes',
    mapColor: '#000000',
    editorColor: options.editorColor ?? '',
    displayText: options.displayText ?? '',
  };
}

function buildBootstrap(): BootstrapResponse {
  const randomRoom = makeArea('random-room', 8, 'p-white');
  randomRoom.Blueprint.Tiles[1][1] = makeTile('p-wall');
  randomRoom.Blueprint.Tiles[1][2] = makeTile('p-wall');
  randomRoom.Blueprint.Tiles[1][3] = makeTile('p-wall');
  randomRoom.Blueprint.Tiles[2][2] = makeTile('p-blue');
  randomRoom.Blueprint.Tiles[2][3] = makeTile('p-blue');
  randomRoom.Blueprint.Tiles[2][4] = makeTile('p-blue');

  const stagedRoom = makeArea('staged-room', 8, 'p-grass');
  stagedRoom.Safe = false;
  stagedRoom.North = 'random-room';
  stagedRoom.South = 'random-room';
  stagedRoom.Transports.push({
    SourceY: 3,
    SourceX: 3,
    DestY: 0,
    DestX: 0,
    DestStage: 'random-room',
    Confirmation: false,
    RejectInteractable: false,
  });
  stagedRoom.Blueprint.Tiles[3][3] = makeTile('p-red', 'i-powder');

  return {
    collections: {
      snaps: {
        Name: 'snaps',
        Spaces: {
          toroid: {
            CollectionName: 'snaps',
            Name: 'toroid',
            Topology: 'plane',
            Latitude: 2,
            Longitude: 2,
            AreaHeight: 8,
            AreaWidth: 8,
            Areas: [randomRoom, stagedRoom],
          },
        },
        Fragments: {
          'base-fragments': [
            {
              id: 'frag-2x2',
              name: 'frag-2x2',
              setName: 'base-fragments',
              blueprint: {
                Tiles: [
                  [makeTile('p-red', 'i-powder'), makeTile('p-round')],
                  [makeTile('p-dark-grass'), makeTile('p-white')],
                ],
                Instructions: [],
                DefaultTileColor: 'white',
                DefaultTileColor1: 'sand',
              },
            },
          ],
        },
        PrototypeSets: {
          'base-prototypes': [
            makePrototype('p-white', { commonName: 'white', walkable: true, cssColor: 'white', ceiling2css: 'white' }),
            makePrototype('p-sand', { commonName: 'sand', walkable: true, cssColor: 'sand', ceiling2css: 'sand' }),
            makePrototype('p-wall', { commonName: 'wall', walkable: false, cssColor: 'wall', ceiling2css: 'wall' }),
            makePrototype('p-grass', { commonName: 'grass', walkable: true, cssColor: 'grass', ceiling2css: 'grass' }),
            makePrototype('p-dark-grass', {
              commonName: 'dark-grass',
              walkable: true,
              cssColor: 'dark-grass',
              ceiling2css: 'dark-grass',
            }),
            makePrototype('p-blue', {
              commonName: 'blue',
              walkable: false,
              cssColor: 'blue',
              ceiling2css: 'blue',
              editorColor: 'green',
            }),
            makePrototype('p-red', {
              commonName: 'red',
              walkable: false,
              cssColor: 'red',
              ceiling2css: 'red',
            }),
            makePrototype('p-round', {
              commonName: 'round-corner',
              walkable: false,
              cssColor: 'round',
              ceiling2css: 'sand r1-{rotate:tl}  r1-{rotate:tr}',
              editorColor: '',
            }),
          ],
        },
        InteractableSets: {
          'base-interactables': [
            {
              id: 'i-powder',
              name: 'PWDR',
              setName: 'base-interactables',
              cssClass: 'powder-blue',
              pushable: true,
              walkable: true,
              fragile: true,
              reactions: 'spark',
              reactionRules: [],
            },
          ],
        },
      },
    },
    colors: [
      { cssClassName: 'white', R: 255, G: 255, B: 255, A: '1' },
      { cssClassName: 'sand', R: 216, G: 196, B: 160, A: '1' },
    ],
  };
}