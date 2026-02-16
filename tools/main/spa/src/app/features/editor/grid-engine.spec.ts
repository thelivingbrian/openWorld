import {
  applyGridTool,
  ensureGround,
  generateMaterials,
  normalizeTile,
} from './grid-engine';
import {
  Blueprint,
  Fragment,
  InteractableDescription,
  Prototype,
  TileData,
} from '../../core/models/editor.models';

describe('grid-engine', () => {
  function makeBlueprint(tiles: TileData[][]): Blueprint {
    return {
      Tiles: tiles,
      Instructions: [],
      DefaultTileColor: 'c0',
      DefaultTileColor1: 'c1',
    };
  }

  function makePrototype(id: string, layer1css = ''): Prototype {
    return {
      id,
      commonName: id,
      cssColor: `${id}-ground`,
      walkable: true,
      layer1css,
      layer2css: '',
      ceiling1css: '',
      ceiling2css: 'base-ceiling',
      setName: 'set-a',
      mapColor: '#000000',
      editorColor: '{rotate:tr}',
      displayText: id,
    };
  }

  function buildInput(blueprint: Blueprint) {
    return {
      blueprint,
      prototypesById: new Map<string, Prototype>(),
      fragmentsById: new Map<string, Fragment>(),
      interactablesById: new Map<string, InteractableDescription>(),
    };
  }

  it('normalizeTile fills missing values', () => {
    const out = normalizeTile({});

    expect(out).toEqual({
      prototypeId: '',
      interactableId: '',
      transformation: { clockwiseRotations: 0 },
    });
  });

  it('ensureGround initializes a ground grid only once', () => {
    const blueprint = makeBlueprint([[{}, {}], [{}, {}]]);
    ensureGround(blueprint);
    const existingGround = blueprint.Ground;

    ensureGround(blueprint);

    expect(existingGround).toBeDefined();
    expect(blueprint.Ground).toBe(existingGround);
    expect(blueprint.Ground).toEqual([
      [{ status: 0 }, { status: 0 }],
      [{ status: 0 }, { status: 0 }],
    ]);
  });

  it('applyGridTool replaces and fills contiguous tiles', () => {
    const blueprint = makeBlueprint([
      [{ prototypeId: 'a' }, { prototypeId: 'a' }, { prototypeId: 'b' }],
      [{ prototypeId: 'a' }, { prototypeId: 'b' }, { prototypeId: 'b' }],
      [{ prototypeId: 'b' }, { prototypeId: 'b' }, { prototypeId: 'a' }],
    ]);
    const maps = buildInput(blueprint);

    applyGridTool({
      ...maps,
      y: 0,
      x: 0,
      tool: 'replace',
      selectedAssetId: 'z',
    });

    expect(blueprint.Tiles[0][0].prototypeId).toBe('z');

    applyGridTool({
      ...maps,
      y: 0,
      x: 1,
      tool: 'fill',
      selectedAssetId: 'x',
    });

    expect(blueprint.Tiles[0][1].prototypeId).toBe('x');
    expect(blueprint.Tiles[1][0].prototypeId).toBe('a');
    expect(blueprint.Tiles[0][2].prototypeId).toBe('b');
  });

  it('applyGridTool fill recurses through vertical and horizontal neighbors', () => {
    const blueprint = makeBlueprint([
      [{ prototypeId: 'a' }, { prototypeId: 'a' }, { prototypeId: 'b' }],
      [{ prototypeId: 'b' }, { prototypeId: 'a' }, { prototypeId: 'b' }],
      [{ prototypeId: 'b' }, { prototypeId: 'a' }, { prototypeId: 'a' }],
    ]);
    const maps = buildInput(blueprint);

    applyGridTool({
      ...maps,
      y: 0,
      x: 0,
      tool: 'fill',
      selectedAssetId: 'z',
    });

    expect(blueprint.Tiles[0][0].prototypeId).toBe('z');
    expect(blueprint.Tiles[0][1].prototypeId).toBe('z');
    expect(blueprint.Tiles[1][1].prototypeId).toBe('z');
    expect(blueprint.Tiles[2][1].prototypeId).toBe('z');
    expect(blueprint.Tiles[2][2].prototypeId).toBe('z');
    expect(blueprint.Tiles[0][2].prototypeId).toBe('b');
    expect(blueprint.Tiles[1][0].prototypeId).toBe('b');
  });

  it('applyGridTool supports between selection rectangle', () => {
    const blueprint = makeBlueprint([
      [{}, {}, {}],
      [{}, {}, {}],
      [{}, {}, {}],
    ]);
    const maps = buildInput(blueprint);

    const selection = applyGridTool({
      ...maps,
      y: 0,
      x: 0,
      tool: 'between',
      selectedAssetId: 'stone',
      selected: { y: 1, x: 2 },
    });

    expect(selection).toEqual({ y: 0, x: 0 });
    expect(blueprint.Tiles[0][0].prototypeId).toBe('stone');
    expect(blueprint.Tiles[1][2].prototypeId).toBe('stone');
    expect(blueprint.Tiles[2][2].prototypeId).toBe('');
  });

  it('applyGridTool places fragment tiles and rotates target cell', () => {
    const blueprint = makeBlueprint([
      [{ prototypeId: 'a' }, { prototypeId: 'a' }],
      [{ prototypeId: 'a' }, { prototypeId: 'a' }],
    ]);
    const fragment: Fragment = {
      id: 'frag-1',
      name: 'frag-1',
      setName: 'set-a',
      blueprint: makeBlueprint([
        [{ prototypeId: 'f1' }, { prototypeId: 'f2' }],
        [{ prototypeId: 'f3' }, { prototypeId: 'f4' }],
      ]),
    };
    const maps = buildInput(blueprint);
    maps.fragmentsById.set(fragment.id, fragment);

    applyGridTool({
      ...maps,
      y: 1,
      x: 1,
      tool: 'place',
      selectedAssetId: 'frag-1',
    });

    expect(blueprint.Tiles[1][1].prototypeId).toBe('f1');
    expect(blueprint.Tiles[0][0].prototypeId).toBe('a');

    applyGridTool({
      ...maps,
      y: 1,
      x: 1,
      tool: 'rotate',
      selectedAssetId: '',
    });

    expect(blueprint.Tiles[1][1].transformation?.clockwiseRotations).toBe(1);
  });

  it('applyGridTool places blueprint instructions and applies source grid', () => {
    const blueprint = makeBlueprint([
      [{}, {}],
      [{}, {}],
    ]);
    const maps = buildInput(blueprint);
    maps.prototypesById.set('proto-1', makePrototype('proto-1'));
    const randomSpy = jest
      .spyOn(crypto, 'randomUUID')
      .mockReturnValue('00000000-0000-4000-8000-000000000001');

    applyGridTool({
      ...maps,
      y: 1,
      x: 0,
      tool: 'place-blueprint',
      selectedAssetId: 'proto-1',
    });

    expect(randomSpy).toHaveBeenCalled();
    expect(blueprint.Instructions.length).toBe(1);
    expect(blueprint.Instructions[0].ID).toBe('00000000-0000-4000-8000-000000000001');
    expect(blueprint.Tiles[1][0].prototypeId).toBe('proto-1');
    randomSpy.mockRestore();
  });

  it('applyGridTool applies pre-existing instructions with 1, 2, and 3 clockwise rotations', () => {
    const blueprint = makeBlueprint([
      [{}, {}, {}, {}, {}, {}, {}, {}],
      [{}, {}, {}, {}, {}, {}, {}, {}],
      [{}, {}, {}, {}, {}, {}, {}, {}],
    ]);
    const maps = buildInput(blueprint);
    maps.fragmentsById.set('frag-rot', {
      id: 'frag-rot',
      name: 'frag-rot',
      setName: 'set-a',
      blueprint: makeBlueprint([
        [{ prototypeId: 'a' }, { prototypeId: 'b', interactableId: 'int' }],
        [{ prototypeId: 'c' }, { prototypeId: 'd' }],
      ]),
    });

    blueprint.Instructions = [
      { ID: 'i1', X: 0, Y: 0, GridAssetId: 'frag-rot', ClockwiseRotations: 1 },
      { ID: 'i2', X: 3, Y: 0, GridAssetId: 'frag-rot', ClockwiseRotations: 2 },
      { ID: 'i3', X: 6, Y: 0, GridAssetId: 'frag-rot', ClockwiseRotations: 3 },
    ];

    applyGridTool({
      ...maps,
      y: 2,
      x: 0,
      tool: 'place-blueprint',
      selectedAssetId: 'missing-asset',
    });

    expect(blueprint.Tiles[0][0].prototypeId).toBe('c');
    expect(blueprint.Tiles[0][1].prototypeId).toBe('a');
    expect(blueprint.Tiles[0][1].interactableId).toBe('');
    expect(blueprint.Tiles[1][0].prototypeId).toBe('d');
    expect(blueprint.Tiles[1][1].prototypeId).toBe('b');
    expect(blueprint.Tiles[1][1].interactableId).toBe('int');

    expect(blueprint.Tiles[0][3].prototypeId).toBe('d');
    expect(blueprint.Tiles[0][4].prototypeId).toBe('c');
    expect(blueprint.Tiles[1][3].prototypeId).toBe('b');
    expect(blueprint.Tiles[1][3].interactableId).toBe('int');
    expect(blueprint.Tiles[1][4].prototypeId).toBe('a');

    expect(blueprint.Tiles[0][6].prototypeId).toBe('b');
    expect(blueprint.Tiles[0][6].interactableId).toBe('int');

    expect(blueprint.Tiles[0][7].prototypeId).toBe('d');
    expect(blueprint.Tiles[1][6].prototypeId).toBe('a');
    expect(blueprint.Tiles[1][7].prototypeId).toBe('c');
  });

  it('applyGridTool sets and clears interactables', () => {
    const blueprint = makeBlueprint([[{}]]);
    const maps = buildInput(blueprint);
    maps.interactablesById.set('door', {
      id: 'door',
      name: 'door',
      setName: 'set-a',
      cssClass: 'door-css',
      pushable: false,
      walkable: true,
      fragile: false,
      reactions: '',
    });

    applyGridTool({
      ...maps,
      y: 0,
      x: 0,
      tool: 'interactable-replace',
      selectedAssetId: 'door',
    });
    expect(blueprint.Tiles[0][0].interactableId).toBe('door');

    applyGridTool({
      ...maps,
      y: 0,
      x: 0,
      tool: 'interactable-delete',
      selectedAssetId: '',
    });
    expect(blueprint.Tiles[0][0].interactableId).toBe('');
  });

  it('applyGridTool toggles single cells and fills connected ground', () => {
    const blueprint = makeBlueprint([
      [{}, {}, {}],
      [{}, {}, {}],
      [{}, {}, {}],
    ]);
    const maps = buildInput(blueprint);

    applyGridTool({
      ...maps,
      y: 1,
      x: 1,
      tool: 'toggle',
      selectedAssetId: '',
    });

    expect(blueprint.Ground?.[1][1].status).toBe(1);

    applyGridTool({
      ...maps,
      y: 0,
      x: 0,
      tool: 'toggle-fill',
      selectedAssetId: '',
    });

    expect(blueprint.Ground?.[0][0].status).toBe(1);
    expect(blueprint.Ground?.[2][2].status).toBe(1);
    expect(blueprint.Ground?.[1][1].status).toBe(1);
  });

  it('toggle smoothCorners handles count === 3 by marking only the zero-status corner', () => {
    const blueprint = makeBlueprint([
      [{}, {}],
      [{}, {}],
    ]);
    blueprint.Ground = [
      [{ status: 0 }, { status: 1 }],
      [{ status: 1 }, { status: 0 }],
    ];
    const maps = buildInput(blueprint);

    applyGridTool({
      ...maps,
      y: 1,
      x: 1,
      tool: 'toggle',
      selectedAssetId: '',
    });

    expect(blueprint.Ground?.[0][0].bottomRight).toBe(true);
    expect(blueprint.Ground?.[0][1].bottomLeft).toBe(false);
    expect(blueprint.Ground?.[1][0].topRight).toBe(false);
    expect(blueprint.Ground?.[1][1].topLeft).toBe(false);
  });

  it('toggle smoothCorners count === 2 diagonal uses Math.random < 0.5 path', () => {
    const blueprint = makeBlueprint([
      [{}, {}],
      [{}, {}],
    ]);
    blueprint.Ground = [
      [{ status: 1 }, { status: 0 }],
      [{ status: 0 }, { status: 0 }],
    ];
    const maps = buildInput(blueprint);
    const randomSpy = jest.spyOn(Math, 'random').mockReturnValue(0.1);

    applyGridTool({
      ...maps,
      y: 1,
      x: 1,
      tool: 'toggle',
      selectedAssetId: '',
    });

    expect(blueprint.Ground?.[0][0].bottomRight).toBe(true);
    expect(blueprint.Ground?.[1][1].topLeft).toBe(true);
    expect(blueprint.Ground?.[0][1].bottomLeft).toBe(false);
    expect(blueprint.Ground?.[1][0].topRight).toBe(false);
    randomSpy.mockRestore();
  });

  it('toggle smoothCorners count === 2 diagonal uses Math.random >= 0.5 path', () => {
    const blueprint = makeBlueprint([
      [{}, {}],
      [{}, {}],
    ]);
    blueprint.Ground = [
      [{ status: 1 }, { status: 0 }],
      [{ status: 0 }, { status: 0 }],
    ];
    const maps = buildInput(blueprint);
    const randomSpy = jest.spyOn(Math, 'random').mockReturnValue(0.9);

    applyGridTool({
      ...maps,
      y: 1,
      x: 1,
      tool: 'toggle',
      selectedAssetId: '',
    });

    expect(blueprint.Ground?.[0][1].bottomLeft).toBe(true);
    expect(blueprint.Ground?.[1][0].topRight).toBe(true);
    expect(blueprint.Ground?.[0][0].bottomRight).toBe(false);
    expect(blueprint.Ground?.[1][1].topLeft).toBe(false);
    randomSpy.mockRestore();
  });

  it('applyGridTool toggle-between toggles a rectangular ground region', () => {
    const blueprint = makeBlueprint([
      [{}, {}, {}, {}],
      [{}, {}, {}, {}],
      [{}, {}, {}, {}],
    ]);
    const maps = buildInput(blueprint);

    const selection = applyGridTool({
      ...maps,
      y: 1,
      x: 2,
      tool: 'toggle-between',
      selectedAssetId: '',
      selected: { y: 0, x: 1 },
    });

    expect(selection).toEqual({ y: 1, x: 2 });
    expect(blueprint.Ground?.[0][1].status).toBe(1);
    expect(blueprint.Ground?.[0][2].status).toBe(1);
    expect(blueprint.Ground?.[1][1].status).toBe(1);
    expect(blueprint.Ground?.[1][2].status).toBe(1);
    expect(blueprint.Ground?.[0][0].status).toBe(0);
    expect(blueprint.Ground?.[2][3].status).toBe(0);

    applyGridTool({
      ...maps,
      y: 1,
      x: 2,
      tool: 'toggle-between',
      selectedAssetId: '',
      selected: { y: 0, x: 1 },
    });

    expect(blueprint.Ground?.[0][1].status).toBe(0);
    expect(blueprint.Ground?.[0][2].status).toBe(0);
    expect(blueprint.Ground?.[1][1].status).toBe(0);
    expect(blueprint.Ground?.[1][2].status).toBe(0);
  });

  it('generateMaterials returns transformed prototype material and ground-only material', () => {
    const blueprint = makeBlueprint([[{ prototypeId: 'proto-1', transformation: { clockwiseRotations: 1 } }]]);
    blueprint.Ground = [[{ status: 1, topLeft: true }]];
    const prototypesById = new Map<string, Prototype>();
    prototypesById.set('proto-1', makePrototype('proto-1', '{rotate:tr}'));

    const full = generateMaterials(blueprint, prototypesById, false);
    const groundOnly = generateMaterials(blueprint, prototypesById, true);

    expect(full[0][0].ground2css).toBe('proto-1-ground');
    expect(full[0][0].layer1css).toBe('br');
    expect(full[0][0].ceiling2css).toBe('br');

    expect(groundOnly[0][0].ground2css).toContain('c1');
    expect(groundOnly[0][0].ground2css).toContain('r0-tl');
    expect(groundOnly[0][0].ground1css).toBe('c0');
  });

  it('generateMaterials appends top-right and bottom corner CSS markers', () => {
    const blueprint = makeBlueprint([[{}]]);
    blueprint.Ground = [[{ status: 0, topRight: true, bottomLeft: true, bottomRight: true }]];

    const groundOnly = generateMaterials(blueprint, new Map<string, Prototype>(), true);

    expect(groundOnly[0][0].ground2css).toContain('r0-tr');
    expect(groundOnly[0][0].ground2css).toContain('r0-bl');
    expect(groundOnly[0][0].ground2css).toContain('r0-br');
    expect(groundOnly[0][0].ground1css).toBe('c1');
  });
});