import { ComponentFixture, TestBed } from '@angular/core/testing';
import { EditorComponent } from './editor.component';
import { EditorApiService } from '../../core/services/editor-api.service';
import { BootstrapResponse, Space } from '../../core/models/editor.models';

describe('EditorComponent', () => {
  let component: EditorComponent;
  let fixture: ComponentFixture<EditorComponent>;
  let api: jasmine.SpyObj<EditorApiService>;

  beforeEach(async () => {
    api = jasmine.createSpyObj<EditorApiService>('EditorApiService', [
      'createCollection',
      'createSpace',
      'createArea',
      'getBootstrap',
      'saveSpace',
      'flattenSpace',
      'savePrototypeSet',
      'saveFragmentSet',
      'saveInteractableSet',
      'saveColors',
      'compile',
      'deploy',
    ]);

    api.getBootstrap.and.callFake(async () => buildBootstrap());
    api.createCollection.and.resolveTo();
    api.createSpace.and.resolveTo();
    api.createArea.and.resolveTo();
    api.saveSpace.and.resolveTo();
    api.flattenSpace.and.resolveTo({ spaceName: 'space-1-flat' });
    api.savePrototypeSet.and.resolveTo();
    api.saveFragmentSet.and.resolveTo();
    api.saveInteractableSet.and.resolveTo();
    api.saveColors.and.resolveTo();
    api.compile.and.resolveTo();
    api.deploy.and.resolveTo();

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

  it('loads bootstrap and initializes editor selections', () => {
    const state = component as any;

    expect(api.getBootstrap).toHaveBeenCalled();
    expect(state.loading()).toBeFalse();
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
    expect(state.showGridTools()).toBeTrue();
    expect(state.showAreaDetails()).toBeFalse();
    expect(state.showTransports()).toBeFalse();
    expect(state.showNeighbors()).toBeFalse();
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

  it('createSpace trims the name and sends payload to API', async () => {
    const state = component as any;
    const current = state.newSpace();
    state.newSpace.set({ ...current, name: '  fresh-space  ' });

    await state.createSpace();

    expect(api.createSpace).toHaveBeenCalledTimes(1);
    expect(api.createSpace).toHaveBeenCalledWith(
      jasmine.objectContaining({
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
    const [, savedSpaceName, savedSpace] = api.saveSpace.calls.mostRecent().args;
    expect(savedSpaceName).toBe('space-1');
    expect((savedSpace as Space).Areas.every((area) => area.Safe)).toBeTrue();
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
              cssClass: '',
              pushable: false,
              walkable: true,
              fragile: false,
              reactions: '',
            },
          ],
        },
        StructureSets: {},
      },
    },
    colors: [{ cssClassName: 'green', R: 0, G: 128, B: 0, A: '1' }],
  };
}
