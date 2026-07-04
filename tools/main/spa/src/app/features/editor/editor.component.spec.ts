import { ComponentFixture, TestBed } from '@angular/core/testing';
import { EditorComponent } from './editor.component';
import { EditorApiService } from '../../core/services/editor-api.service';
import { BootstrapResponse, Space } from '../../core/models/editor.models';

describe('EditorComponent', () => {
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
      flattenSpace: jest.fn().mockResolvedValue({ spaceName: 'space-1-flat' }),
      savePrototypeSet: jest.fn().mockResolvedValue(undefined),
      saveFragmentSet: jest.fn().mockResolvedValue(undefined),
      saveInteractableSet: jest.fn().mockResolvedValue(undefined),
      saveColors: jest.fn().mockResolvedValue(undefined),
      compile: jest.fn().mockResolvedValue(undefined),
      deploy: jest.fn().mockResolvedValue(undefined),
    } as unknown as jest.Mocked<EditorApiService>;

    await TestBed.configureTestingModule({
      imports: [EditorComponent],
      providers: [
        { provide: EditorApiService, useValue: api },
      ],
    })
    .compileComponents();

    fixture = TestBed.createComponent(EditorComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('shows only the operations supported by the active editor context', () => {
    const text = () => fixture.nativeElement.querySelector('.topbar').textContent;
    const buttons = () => Array.from(
      fixture.nativeElement.querySelectorAll('.topbar button') as NodeListOf<HTMLButtonElement>,
      (button: HTMLButtonElement) => button.textContent?.trim(),
    );

    expect(buttons()).toEqual(expect.arrayContaining(['Compile', 'Deploy', 'New']));
    expect(buttons()).not.toEqual(expect.arrayContaining(['Publish', 'Launch']));

    (component as any).hosted = true;
    fixture.detectChanges();

    expect(buttons()).toEqual(expect.arrayContaining(['Publish', 'Launch']));
    expect(buttons()).not.toEqual(expect.arrayContaining(['Compile', 'Deploy', 'New']));
    expect(text()).toContain('Worlds');
    expect(Array.from(fixture.nativeElement.querySelectorAll('button')).some((button: any) => button.textContent.includes('Flatten Space'))).toBe(false);
  });

  it('loads palette colors without importing the host application stylesheet', () => {
    const stylesheetHrefs = Array.from(
      document.querySelectorAll<HTMLLinkElement>('link[rel="stylesheet"]'),
      (link) => link.getAttribute('href'),
    );

    expect(stylesheetHrefs).toContain('/assets/colors.css');
    expect(stylesheetHrefs).not.toContain('/assets/style.css');
  });

  it('loads bootstrap and initializes editor selections', () => {
    const state = component as any;

    expect(api.getBootstrap).toHaveBeenCalled();
    expect(state.loading()).toBe(false);
    expect(state.collectionName()).toBe('alpha');
    expect(state.spaceName()).toBe('space-1');
    expect(state.modifySpaceName()).toBe('space-1');
    expect(state.areaName()).toBe('space-1:0-0');
    expect(state.prototypeSet()).toBe('base-prototypes');
    expect(state.fragmentSet()).toBe('base-fragments');
    expect(state.interactableSet()).toBe('base-interactables');
    expect(state.selectedAssetId()).toBe('proto-1');
    expect(state.tool()).toBe('select');
  });

  it('setViewMode resets area edit panels when leaving world view', () => {
    const state = component as any;
    state.showGridTools.set(false);
    state.showAreaDetails.set(true);
    state.showTransports.set(true);
    state.showNeighbors.set(true);
    state.selection.set({ y: 1, x: 1 });

    state.setViewMode('create');

    expect(state.viewMode()).toBe('create');
    expect(state.gridTarget()).toBe('area');
    expect(state.showGridTools()).toBe(true);
    expect(state.showAreaDetails()).toBe(false);
    expect(state.showTransports()).toBe(false);
    expect(state.showNeighbors()).toBe(false);
    expect(state.selection()).toBeUndefined();
  });

  it('setViewMode switches grid target to fragment in fragments mode', () => {
    const state = component as any;

    state.setViewMode('fragments');

    expect(state.gridTarget()).toBe('fragment');
  });

  it('navigates to an area in another space', () => {
    const state = component as any;

    state.spaceName.set('space-1');
    state.areaName.set('space-1:0-0');
    state.goToArea('target-area');

    expect(state.spaceName()).toBe('space-2');
    expect(state.areaName()).toBe('target-area');
  });

  it('adds, duplicates, and deletes transports on current area', () => {
    const state = component as any;
    const area = state.currentArea();

    expect(area.Transports.length).toBe(1);

    state.addTransport();
    expect(area.Transports.length).toBe(2);

    state.duplicateTransport(0);
    expect(area.Transports.length).toBe(3);
    expect(area.Transports[2]).toEqual(area.Transports[0]);

    state.deleteTransport(1);
    expect(area.Transports.length).toBe(2);
  });

  it('onFixtureChange sets the expected tool and selected asset', () => {
    const state = component as any;

    state.fixture.set('fragment');
    state.onFixtureChange();
    expect(state.tool()).toBe('place-blueprint');
    expect(state.selectedAssetId()).toBe('frag-1');

    state.fixture.set('ground');
    state.onFixtureChange();
    expect(state.tool()).toBe('toggle');
    expect(state.selectedAssetId()).toBe('');
  });

  it('onGridClick expands selected information when a tile is selected', () => {
    const state = component as any;
    state.showSelectedInformation.set(false);

    state.onGridClick(0, 0);

    expect(state.selection()).toEqual({ y: 0, x: 0 });
    expect(state.showSelectedInformation()).toBe(true);
  });

  it('onGridClick does not expand selected information when no tile selection is created', () => {
    const state = component as any;
    state.showSelectedInformation.set(false);
    state.selection.set(undefined);
    state.tool.set('replace');

    state.onGridClick(0, 0);

    expect(state.selection()).toBeUndefined();
    expect(state.showSelectedInformation()).toBe(false);
  });

  it('selectedTile reads from area blueprint by default and fragment blueprint in fragments mode', () => {
    const state = component as any;
    const area = state.currentArea();
    area.Blueprint.Tiles = [[{ prototypeId: 'proto-1' }]];
    state.selection.set({ y: 0, x: 0 });

    expect(state.selectedTile()?.prototypeId).toBe('proto-1');

    state.setViewMode('fragments');
    state.selection.set({ y: 0, x: 0 });
    state.editedFragment().blueprint.Tiles = [[{ prototypeId: 'frag-proto' }]];

    expect(state.selectedTile()?.prototypeId).toBe('frag-proto');
  });

  it('selectedTile returns undefined when selection is out of row bounds', () => {
    const state = component as any;
    const area = state.currentArea();
    area.Blueprint.Tiles = [[{ prototypeId: 'proto-1' }]];
    state.selection.set({ y: 0, x: 2 });

    expect(state.selectedTile()).toBeUndefined();
  });

  it('selected tile asset computeds resolve trimmed prototype and interactable ids', () => {
    const state = component as any;
    const area = state.currentArea();
    area.Blueprint.Tiles = [[{ prototypeId: '  proto-1  ', interactableId: '  inter-1  ' }]];
    state.selection.set({ y: 0, x: 0 });

    expect(state.selectedTilePrototype()?.id).toBe('proto-1');
    expect(state.selectedTileInteractable()?.id).toBe('inter-1');
  });

  it('selectedTileInstructions filters by selected coordinate and maps known/unknown assets', () => {
    const state = component as any;
    const area = state.currentArea();
    area.Blueprint.Instructions = [
      { ID: 'i-1', X: 0, Y: 0, GridAssetId: 'proto-1', ClockwiseRotations: 0 },
      { ID: 'i-2', X: 0, Y: 0, GridAssetId: 'missing-asset', ClockwiseRotations: 0 },
      { ID: 'i-3', X: 1, Y: 0, GridAssetId: 'proto-1', ClockwiseRotations: 0 },
    ];
    state.selection.set({ y: 0, x: 0 });

    const infos = state.selectedTileInstructions();
    expect(infos).toHaveLength(2);
    expect(infos[0]).toEqual(
      expect.objectContaining({
        index: 0,
        assetType: 'prototype',
        assetLabel: 'Prototype: stone',
      }),
    );
    expect(infos[1]).toEqual(
      expect.objectContaining({
        index: 1,
        assetLabel: 'Unknown asset (missing-asset)',
      }),
    );
    expect(infos[1]).not.toHaveProperty('assetType');
  });

  it('createSpace trims the name and sends payload to API', async () => {
    const state = component as any;
    const current = state.newSpace();
    state.newSpace.set({ ...current, name: '  fresh-space  ' });

    await state.createSpace();

    expect(api.createSpace).toHaveBeenCalledTimes(1);
    expect(api.createSpace).toHaveBeenCalledWith(
      expect.objectContaining({
        collectionName: 'alpha',
        name: 'fresh-space',
      }),
    );
    expect(state.status()).toBe('Space created.');
  });

  it('applyAreaPropertyToSpace updates all areas and persists the space', async () => {
    const state = component as any;

    state.modifySpaceName.set('space-1');
    state.bulkAreaProperty.set('safe');
    state.bulkAreaValueBoolean.set(true);

    await state.applyAreaPropertyToSpace();

    expect(api.saveSpace).toHaveBeenCalledTimes(1);
    const [, savedSpaceName, savedSpace] = api.saveSpace.mock.calls.at(-1)!;
    expect(savedSpaceName).toBe('space-1');
    expect((savedSpace as Space).Areas.every((area) => area.Safe)).toBe(true);
    expect(state.status()).toBe('Updated 2 areas in space-1.');
  });

  it('flattenSpace blocks unsupported topologies', async () => {
    const state = component as any;
    state.spaceName.set('space-2');

    await state.flattenSpace();

    expect(api.flattenSpace).not.toHaveBeenCalled();
    expect(state.status()).toBe('Only simply tiled spaces may be flattened.');
  });

  it('flattenSpace calls API and updates selected space on success', async () => {
    const state = component as any;
    state.spaceName.set('space-1');

    await state.flattenSpace();

    expect(api.flattenSpace).toHaveBeenCalledWith('alpha', 'space-1');
    expect(state.spaceName()).toBe('space-1-flat');
    expect(state.modifySpaceName()).toBe('space-1-flat');
    expect(state.status()).toBe('Space flattened into space-1-flat.');
  });

  it('resetUnsavedChanges reloads current space and discards local edits', async () => {
    const state = component as any;
    const area = state.currentArea();
    area.Blueprint.DefaultTileColor = 'purple';
    state.instructionEditedIds.set({ instructionA: true });
    state.touchBootstrap();

    await state.resetUnsavedChanges();

    expect(api.getBootstrap).toHaveBeenCalledTimes(2);
    expect(state.spaceName()).toBe('space-1');
    expect(state.areaName()).toBe('space-1:0-0');
    expect(state.currentArea().Blueprint.DefaultTileColor).toBe('green');
    expect(state.instructionEditedIds()).toEqual({});
    expect(state.status()).toBe('Unsaved changes in space-1 reset.');
  });

  it('addInteractableSet adds a new set, selects it, and reports success', () => {
    const state = component as any;

    state.newInteractableSetName.set('advanced');
    state.addInteractableSet();

    expect(state.currentCollection().InteractableSets.advanced).toEqual([]);
    expect(state.interactableSets()).toEqual(expect.arrayContaining(['advanced']));
    expect(state.interactableSet()).toBe('advanced');
    expect(state.newInteractableSetName()).toBe('');
    expect(state.status()).toBe('Interactable set "advanced" added.');
  });

  it('addInteractable selects a new interactable and initializes editable default state', () => {
    const state = component as any;

    state.interactableSet.set('base-interactables');
    state.onInteractableSetChange();
    state.interactableStateEditName.set('missing-state');

    state.addInteractable();

    expect(state.editedInteractable()?.name).toBe('new-interactable');
    expect(state.interactableStateEditName()).toBe('default');
    expect(state.editedInteractableState()).toEqual(
      expect.objectContaining({
        cssClass: '',
        pushable: false,
        walkable: false,
        fragile: false,
        stickyGroups: [],
      }),
    );
  });

  it('addInteractable immediately updates a newly created interactable set without switching sets', () => {
    const state = component as any;

    state.newInteractableSetName.set('example');
    state.addInteractableSet();
    expect(state.interactableSet()).toBe('example');
    expect(state.interactables()).toHaveLength(0);

    state.addInteractable();

    expect(state.interactables()).toHaveLength(1);
    expect(state.interactables()[0].name).toBe('new-interactable');
    expect(state.interactableEditId()).toBe(state.interactables()[0].id);
    expect(state.editedInteractable()?.id).toBe(state.interactables()[0].id);
  });

  it('can rename an edited interactable state', () => {
    const state = component as any;

    state.interactableSet.set('base-interactables');
    state.onInteractableSetChange();
    state.addInteractable();
    state.addEditedInteractableState();
    expect(state.editedInteractableStateNames()).toEqual(expect.arrayContaining(['state-1']));

    state.interactableStateRenameName.set('armed');
    state.renameEditedInteractableState();

    expect(state.editedInteractableStateNames()).toEqual(expect.arrayContaining(['default', 'armed']));
    expect(state.editedInteractableStateNames()).not.toEqual(expect.arrayContaining(['state-1']));
    expect(state.interactableStateEditName()).toBe('armed');
    expect(state.interactableStateRenameName()).toBe('armed');
    expect(state.status()).toBe('State renamed to "armed".');
  });

  it('addEditedInteractableState clones reaction rules instead of sharing references', () => {
    const state = component as any;

    state.interactableSet.set('base-interactables');
    state.onInteractableSetChange();
    state.addInteractable();

    const interactable = state.editedInteractable();
    interactable.states.default.reactionRules = [
      {
        reactsWith: 'everything',
        reactsWithArgs: ['a'],
        reaction: 'pass',
        reactionArgs: ['b'],
      },
    ];

    state.addEditedInteractableState();
    const newStateName = state.interactableStateEditName();
    const defaultRules = interactable.states.default.reactionRules;
    const newRules = interactable.states[newStateName].reactionRules;

    expect(newRules).not.toBe(defaultRules);
    expect(newRules[0]).not.toBe(defaultRules[0]);

    newRules[0].reaction = 'eat';
    newRules[0].reactionArgs.push('c');

    expect(interactable.states.default.reactionRules[0].reaction).toBe('pass');
    expect(interactable.states.default.reactionRules[0].reactionArgs).toEqual(['b']);
  });

  it('addInteractableSet reports validation errors for blank and duplicate names', () => {
    const state = component as any;

    state.newInteractableSetName.set('   ');
    state.addInteractableSet();
    expect(state.status()).toBe('Interactable set name cannot be blank.');

    state.newInteractableSetName.set('base-interactables');
    state.addInteractableSet();
    expect(state.status()).toBe('Interactable set "base-interactables" already exists.');
  });

  it('gridLayerDynamicStyle returns empty object in world view and computes style in fragment view', () => {
    const state = component as any;

    // In world view (default), dynamic tokens produce a static snapshot (non-animated, no signal dep)
    state.viewMode.set('world');
    state.dynamicTimeMs.set(0);
    const worldStyle = state.gridLayerDynamicStyle('rainbow', 0, 0);
    state.dynamicTimeMs.set(99999);
    const worldStyleLater = state.gridLayerDynamicStyle('rainbow', 0, 0);
    // In world view time is fixed at 0 so the result is the same regardless of dynamicTimeMs
    expect(worldStyle).toEqual(worldStyleLater);
    expect(worldStyle['backgroundColor']).toBeDefined();

    // In fragment view, the style changes as dynamicTimeMs advances
    state.viewMode.set('fragments');
    state.dynamicTimeMs.set(0);
    const fragStyleT0 = state.gridLayerDynamicStyle('rainbow', 0, 0);
    state.dynamicTimeMs.set(10000);
    const fragStyleT1 = state.gridLayerDynamicStyle('rainbow', 0, 0);
    expect(fragStyleT0['backgroundColor']).toBeDefined();
    expect(fragStyleT0['backgroundColor']).not.toBe(fragStyleT1['backgroundColor']);
  });

  it('gridLayerDynamicStyle returns empty object for static class strings in all view modes', () => {
    const state = component as any;

    state.viewMode.set('world');
    expect(state.gridLayerDynamicStyle('blue green', 0, 0)).toEqual({});

    state.viewMode.set('fragments');
    expect(state.gridLayerDynamicStyle('blue green', 0, 0)).toEqual({});
  });

  it('prototypePreviewCeiling2Style uses editorColor when toggle is active and ceiling2css otherwise', () => {
    const state = component as any;

    const proto = { editorColor: 'rainbow', ceiling2css: 'water' };

    state.prototypePreviewUseEditorColor.set(true);
    const styleWithEditorColor = state.prototypePreviewCeiling2Style(proto);
    expect(styleWithEditorColor['backgroundColor']).toBeDefined(); // rainbow → backgroundColor

    state.prototypePreviewUseEditorColor.set(false);
    const styleWithCeiling2 = state.prototypePreviewCeiling2Style(proto);
    expect(styleWithCeiling2['backgroundColor']).toBeDefined(); // water → backgroundColor

    // They should differ since rainbow and water generate different colors at t=0
    state.dynamicTimeMs.set(0);
    state.prototypePreviewUseEditorColor.set(true);
    const rainbowStyle = state.prototypePreviewCeiling2Style({ editorColor: 'rainbow', ceiling2css: 'green' });
    state.prototypePreviewUseEditorColor.set(false);
    const waterStyle = state.prototypePreviewCeiling2Style({ editorColor: 'rainbow', ceiling2css: 'water' });
    expect(rainbowStyle['backgroundColor']).not.toBe(waterStyle['backgroundColor']);
  });

  it('prototypePreviewCeiling2Style returns empty object for undefined prototype', () => {
    const state = component as any;
    expect(state.prototypePreviewCeiling2Style(undefined)).toEqual({});
  });
});

function buildBootstrap(): BootstrapResponse {
  return {
    collections: {
      alpha: {
        Name: 'alpha',
        Spaces: {
          'space-1': {
            CollectionName: 'alpha',
            Name: 'space-1',
            Topology: 'plane',
            Latitude: 2,
            Longitude: 2,
            AreaHeight: 2,
            AreaWidth: 2,
            Areas: [
              {
                Name: 'space-1:0-0',
                Safe: false,
                Blueprint: {
                  Tiles: [[{ prototypeId: 'proto-1' }]],
                  Instructions: [],
                  DefaultTileColor: 'green',
                  DefaultTileColor1: 'brown',
                },
                Transports: [
                  {
                    SourceY: 0,
                    SourceX: 0,
                    DestY: 1,
                    DestX: 1,
                    DestStage: 'target-area',
                    Confirmation: false,
                    RejectInteractable: false,
                  },
                ],
                North: '',
                South: '',
                East: '',
                West: '',
                Weather: '',
                LoadStrategy: '',
                SpawnStrategy: '',
                BroadcastGroup: '',
              },
              {
                Name: 'space-1:0-1',
                Safe: false,
                Blueprint: {
                  Tiles: [[{ prototypeId: 'proto-1' }]],
                  Instructions: [],
                  DefaultTileColor: 'green',
                  DefaultTileColor1: 'brown',
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
              },
            ],
          },
          'space-2': {
            CollectionName: 'alpha',
            Name: 'space-2',
            Topology: 'hex',
            Latitude: 1,
            Longitude: 1,
            AreaHeight: 1,
            AreaWidth: 1,
            Areas: [
              {
                Name: 'target-area',
                Safe: false,
                Blueprint: {
                  Tiles: [[{ prototypeId: 'proto-1' }]],
                  Instructions: [],
                  DefaultTileColor: 'gray',
                  DefaultTileColor1: 'black',
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
              },
            ],
          },
        },
        Fragments: {
          'base-fragments': [
            {
              id: 'frag-1',
              name: 'fragment-one',
              setName: 'base-fragments',
              blueprint: {
                Tiles: [[{ prototypeId: 'proto-1' }]],
                Instructions: [],
                DefaultTileColor: 'green',
                DefaultTileColor1: 'brown',
              },
            },
          ],
        },
        PrototypeSets: {
          'base-prototypes': [
            {
              id: 'proto-1',
              commonName: 'stone',
              cssColor: '',
              walkable: true,
              layer1css: '',
              layer2css: '',
              ceiling1css: '',
              ceiling2css: '',
              setName: 'base-prototypes',
              mapColor: '',
              editorColor: '',
              displayText: '',
            },
          ],
        },
        InteractableSets: {
          'base-interactables': [
            {
              id: 'inter-1',
              name: 'switch',
              setName: 'base-interactables',
              state: '',
              cssClass: '',
              pushable: false,
              walkable: true,
              fragile: false,
              reactions: '',
              reactionRules: [],
            },
          ],
        },
      },
    },
    colors: [{ cssClassName: 'green', R: 0, G: 128, B: 0, A: '1' }],
  };
}
