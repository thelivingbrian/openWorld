import { Fragment, GridSelection, InteractableDescription, Material, Prototype, TileData, Transformation, Blueprint, Cell, Instruction } from './editor.models';

export type Tool =
  | 'select'
  | 'replace'
  | 'fill'
  | 'between'
  | 'place'
  | 'rotate'
  | 'place-blueprint'
  | 'interactable-select'
  | 'interactable-replace'
  | 'interactable-delete'
  | 'toggle'
  | 'toggle-between'
  | 'toggle-fill';

export function normalizeTile(tile: TileData): TileData {
  return {
    prototypeId: tile.prototypeId ?? '',
    interactableId: tile.interactableId ?? '',
    transformation: { clockwiseRotations: tile.transformation?.clockwiseRotations ?? 0 }
  };
}

export function ensureGround(blueprint: Blueprint): void {
  if (blueprint.Ground && blueprint.Ground.length > 0) {
    return;
  }
  blueprint.Ground = blueprint.Tiles.map((row) => row.map(() => ({ status: 0 })));
}

function mod(input: number, n: number): number {
  return ((input % n) + n) % n;
}

function rotateCss(input: string, clockwiseRotations: number): string {
  const options = ['tr', 'br', 'bl', 'tl'];
  const idx = options.findIndex((entry) => entry === input);
  if (idx < 0) {
    return input;
  }
  return options[mod(idx + clockwiseRotations, 4)];
}

function transformCss(input: string, transformation: Transformation): string {
  return (input ?? '').replace(/\{([^:]*):([^}]*)\}/g, (_m, key: string, value: string) => {
    if (key === 'rotate') {
      return rotateCss(value, transformation.clockwiseRotations ?? 0);
    }
    return value;
  });
}

function applyTransformForEditor(proto: Prototype, transformation: Transformation): Material {
  const ceiling2 = proto.editorColor || proto.ceiling2css;
  return {
    walkable: proto.walkable,
    ground2css: proto.cssColor,
    layer1css: transformCss(proto.layer1css, transformation),
    layer2css: transformCss(proto.layer2css, transformation),
    ceiling1css: transformCss(proto.ceiling1css, transformation),
    ceiling2css: transformCss(ceiling2, transformation),
    displayText: proto.displayText
  };
}

function addGroundToMaterial(material: Material, cell: Cell | undefined, color0: string, color1: string): Material {
  const out: Material = { ...material };
  if (!cell) {
    out.ground1css = color0;
    return out;
  }
  if (out.ground2css) {
    return out;
  }

  let primary = color0;
  let secondary = color1;
  if (cell.status !== 0) {
    primary = color1;
    secondary = color0;
  }
  out.ground2css = primary;
  if (cell.topLeft || cell.topRight || cell.bottomLeft || cell.bottomRight) {
    out.ground1css = secondary;
  }
  if (cell.topLeft) {
    out.ground2css += ' r0-tl';
  }
  if (cell.topRight) {
    out.ground2css += ' r0-tr';
  }
  if (cell.bottomLeft) {
    out.ground2css += ' r0-bl';
  }
  if (cell.bottomRight) {
    out.ground2css += ' r0-br';
  }
  return out;
}

function smoothCorners(cells: Cell[][]): void {
  for (let y = 0; y < cells.length - 1; y += 1) {
    for (let x = 0; x < cells[y].length - 1; x += 1) {
      const a = cells[y][x];
      const b = cells[y][x + 1];
      const c = cells[y + 1][x];
      const d = cells[y + 1][x + 1];
      const count = a.status + b.status + c.status + d.status;

      a.bottomRight = false;
      b.bottomLeft = false;
      c.topRight = false;
      d.topLeft = false;

      if (count === 4 || count === 0) {
        continue;
      }
      if (count === 3) {
        a.bottomRight = a.status !== 1;
        b.bottomLeft = b.status !== 1;
        c.topRight = c.status !== 1;
        d.topLeft = d.status !== 1;
        continue;
      }
      if (count === 1) {
        a.bottomRight = a.status === 1;
        b.bottomLeft = b.status === 1;
        c.topRight = c.status === 1;
        d.topLeft = d.status === 1;
        continue;
      }
      if (count === 2) {
        if (a.status === b.status || a.status === c.status) {
          continue;
        }
        if (Math.random() < 0.5) {
          a.bottomRight = true;
          d.topLeft = true;
        } else {
          b.bottomLeft = true;
          c.topRight = true;
        }
      }
    }
  }
}

function subGrid(grid: Cell[][], y: number, x: number, height: number, width: number): Cell[][] {
  const rowStart = Math.max(0, y);
  const rowEnd = Math.min(y + height, grid.length);
  const colStart = Math.max(0, x);
  const colEnd = Math.min(x + width, grid[0].length);

  const result: Cell[][] = [];
  for (let row = rowStart; row < rowEnd; row += 1) {
    result.push(grid[row].slice(colStart, colEnd));
  }
  return result;
}

function fill(y: number, x: number, grid: TileData[][], selectedPrototypeId: string, seen: boolean[][], targetId: string): void {
  seen[y][x] = true;
  grid[y][x].prototypeId = selectedPrototypeId;

  const deltas = [-1, 1];
  for (const delta of deltas) {
    if (y + delta >= 0 && y + delta < grid.length) {
      const shouldFill = !seen[y + delta][x] && (grid[y + delta][x].prototypeId ?? '') === targetId;
      if (shouldFill) {
        fill(y + delta, x, grid, selectedPrototypeId, seen, targetId);
      }
    }
    if (x + delta >= 0 && x + delta < grid[y].length) {
      const shouldFill = !seen[y][x + delta] && (grid[y][x + delta].prototypeId ?? '') === targetId;
      if (shouldFill) {
        fill(y, x + delta, grid, selectedPrototypeId, seen, targetId);
      }
    }
  }
}

function toggleFill(y: number, x: number, grid: Cell[][], seen: boolean[][], selectedStatus: number): void {
  seen[y][x] = true;
  grid[y][x].status = (grid[y][x].status + 1) % 2;

  const deltas = [-1, 1];
  for (const delta of deltas) {
    if (y + delta >= 0 && y + delta < grid.length) {
      const shouldToggle = !seen[y + delta][x] && grid[y + delta][x].status === selectedStatus;
      if (shouldToggle) {
        toggleFill(y + delta, x, grid, seen, selectedStatus);
      }
    }
    if (x + delta >= 0 && x + delta < grid[y].length) {
      const shouldToggle = !seen[y][x + delta] && grid[y][x + delta].status === selectedStatus;
      if (shouldToggle) {
        toggleFill(y, x + delta, grid, seen, selectedStatus);
      }
    }
  }
}

function rotateClockwise<T>(input: T[][]): T[][] {
  const outputHeight = input[0].length;
  const out: T[][] = Array.from({ length: outputHeight }, () => Array.from({ length: input.length }));
  for (let y = 0; y < outputHeight; y += 1) {
    for (let x = 0; x < input.length; x += 1) {
      out[y][x] = input[input.length - x - 1][y];
    }
  }
  return out;
}

function rotateTimesN(input: TileData[][], n: number): TileData[][] {
  let out = input.map((row) => row.map((tile) => normalizeTile(tile)));
  for (let i = 0; i < mod(n, 4); i += 1) {
    out = rotateClockwise(out);
    for (const row of out) {
      for (const tile of row) {
        tile.transformation = { clockwiseRotations: (tile.transformation?.clockwiseRotations ?? 0) + 1 };
      }
    }
  }
  return out;
}

function pasteTiles(y: number, x: number, source: TileData[][], dest: TileData[][]): void {
  for (let row = 0; row < dest.length; row += 1) {
    if (y + row >= source.length) {
      break;
    }
    for (let col = 0; col < dest[row].length; col += 1) {
      if (x + col >= source[y + row].length) {
        break;
      }
      const tile = dest[row][col];
      if (tile.prototypeId) {
        source[y + row][x + col].prototypeId = tile.prototypeId;
        source[y + row][x + col].transformation = tile.transformation ?? { clockwiseRotations: 0 };
      }
      if (tile.interactableId) {
        source[y + row][x + col].interactableId = tile.interactableId;
      }
    }
  }
}

function getTileGridByAssetId(assetId: string, prototypesById: Map<string, Prototype>, fragmentsById: Map<string, Fragment>): TileData[][] {
  const fragment = fragmentsById.get(assetId);
  if (fragment) {
    return fragment.blueprint.Tiles;
  }
  const proto = prototypesById.get(assetId);
  if (proto) {
    return [[{ prototypeId: assetId, transformation: { clockwiseRotations: 0 } }]];
  }
  return [];
}

function applyInstruction(tiles: TileData[][], instruction: Instruction, prototypesById: Map<string, Prototype>, fragmentsById: Map<string, Fragment>): void {
  const grid = rotateTimesN(getTileGridByAssetId(instruction.GridAssetId, prototypesById, fragmentsById), instruction.ClockwiseRotations);
  pasteTiles(instruction.Y, instruction.X, tiles, grid);
}

function applyEveryInstruction(blueprint: Blueprint, prototypesById: Map<string, Prototype>, fragmentsById: Map<string, Fragment>): void {
  for (const instruction of blueprint.Instructions) {
    applyInstruction(blueprint.Tiles, instruction, prototypesById, fragmentsById);
  }
}

export function applyGridTool(input: {
  y: number;
  x: number;
  tool: Tool;
  selectedAssetId: string;
  selected?: GridSelection;
  blueprint: Blueprint;
  prototypesById: Map<string, Prototype>;
  fragmentsById: Map<string, Fragment>;
  interactablesById: Map<string, InteractableDescription>;
}): GridSelection | undefined {
  const { y, x, tool, selectedAssetId, blueprint, prototypesById, fragmentsById, interactablesById } = input;
  const selected = input.selected;

  for (const row of blueprint.Tiles) {
    for (let i = 0; i < row.length; i += 1) {
      row[i] = normalizeTile(row[i]);
    }
  }

  switch (tool) {
    case 'select':
      return { y, x };

    case 'replace': {
      blueprint.Tiles[y][x].prototypeId = selectedAssetId;
      return selected;
    }

    case 'fill': {
      const targetId = blueprint.Tiles[y][x].prototypeId ?? '';
      const seen = blueprint.Tiles.map((row) => row.map(() => false));
      fill(y, x, blueprint.Tiles, selectedAssetId, seen, targetId);
      return selected;
    }

    case 'between': {
      const anchor = selected ?? { y, x };
      const lowY = Math.min(y, anchor.y);
      const highY = Math.max(y, anchor.y);
      const lowX = Math.min(x, anchor.x);
      const highX = Math.max(x, anchor.x);
      for (let row = lowY; row <= highY; row += 1) {
        for (let col = lowX; col <= highX; col += 1) {
          blueprint.Tiles[row][col].prototypeId = selectedAssetId;
        }
      }
      return { y, x };
    }

    case 'place': {
      const fragment = fragmentsById.get(selectedAssetId);
      if (!fragment) {
        return selected;
      }
      for (let row = 0; row < fragment.blueprint.Tiles.length; row += 1) {
        if (y + row >= blueprint.Tiles.length) {
          break;
        }
        for (let col = 0; col < fragment.blueprint.Tiles[row].length; col += 1) {
          if (x + col >= blueprint.Tiles[y + row].length) {
            break;
          }
          blueprint.Tiles[y + row][x + col] = normalizeTile(fragment.blueprint.Tiles[row][col]);
        }
      }
      return selected;
    }

    case 'rotate': {
      const tx = blueprint.Tiles[y][x].transformation ?? { clockwiseRotations: 0 };
      tx.clockwiseRotations = mod((tx.clockwiseRotations ?? 0) + 1, 4);
      blueprint.Tiles[y][x].transformation = tx;
      return selected;
    }

    case 'place-blueprint': {
      if (!selectedAssetId) {
        return selected;
      }
      blueprint.Instructions = blueprint.Instructions ?? [];
      blueprint.Instructions.push({
        ID: crypto.randomUUID(),
        X: x,
        Y: y,
        GridAssetId: selectedAssetId,
        ClockwiseRotations: 0
      });
      applyEveryInstruction(blueprint, prototypesById, fragmentsById);
      return selected;
    }

    case 'interactable-replace': {
      const interactable = interactablesById.get(selectedAssetId);
      blueprint.Tiles[y][x].interactableId = interactable?.id ?? '';
      return selected;
    }

    case 'interactable-delete': {
      blueprint.Tiles[y][x].interactableId = '';
      return selected;
    }

    case 'toggle': {
      ensureGround(blueprint);
      const ground = blueprint.Ground!;
      ground[y][x].status = (ground[y][x].status + 1) % 2;
      smoothCorners(subGrid(ground, y - 1, x - 1, 3, 3));
      return selected;
    }

    case 'toggle-between': {
      ensureGround(blueprint);
      const ground = blueprint.Ground!;
      const anchor = selected ?? { y, x };
      const lowY = Math.min(y, anchor.y);
      const highY = Math.max(y, anchor.y);
      const lowX = Math.min(x, anchor.x);
      const highX = Math.max(x, anchor.x);
      for (let row = lowY; row <= highY; row += 1) {
        for (let col = lowX; col <= highX; col += 1) {
          ground[row][col].status = (ground[row][col].status + 1) % 2;
        }
      }
      smoothCorners(subGrid(ground, lowY - 1, lowX - 1, highY - lowY + 3, highX - lowX + 3));
      return { y, x };
    }

    case 'toggle-fill': {
      ensureGround(blueprint);
      const ground = blueprint.Ground!;
      const selectedStatus = ground[y][x].status;
      const seen = ground.map((row) => row.map(() => false));
      toggleFill(y, x, ground, seen, selectedStatus);
      smoothCorners(ground);
      return selected;
    }

    default:
      return selected;
  }
}

export function generateMaterials(
  blueprint: Blueprint,
  prototypesById: Map<string, Prototype>,
  groundOnly: boolean
): Material[][] {
  ensureGround(blueprint);
  const out: Material[][] = [];

  for (let y = 0; y < blueprint.Tiles.length; y += 1) {
    const row: Material[] = [];
    for (let x = 0; x < blueprint.Tiles[y].length; x += 1) {
      const cell = blueprint.Ground?.[y]?.[x];
      if (groundOnly) {
        row.push(addGroundToMaterial({}, cell, blueprint.DefaultTileColor, blueprint.DefaultTileColor1));
        continue;
      }

      const tile = normalizeTile(blueprint.Tiles[y][x]);
      const proto = prototypesById.get(tile.prototypeId ?? '');
      let material: Material = { ground2css: 'blue', layer1css: 'green red-b thick' };
      if (proto) {
        material = applyTransformForEditor(proto, tile.transformation ?? { clockwiseRotations: 0 });
      }
      row.push(addGroundToMaterial(material, cell, blueprint.DefaultTileColor, blueprint.DefaultTileColor1));
    }
    out.push(row);
  }

  return out;
}
