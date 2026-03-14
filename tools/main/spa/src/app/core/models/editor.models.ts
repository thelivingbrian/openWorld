export interface Color {
  cssClassName: string;
  R: number;
  G: number;
  B: number;
  A: string;
}

export interface Transformation {
  clockwiseRotations?: number;
}

export interface TileData {
  prototypeId?: string;
  transformation?: Transformation;
  interactableId?: string;
  interactableState?: string;
}

export interface Cell {
  status: number;
  bottomRight?: boolean;
  bottomLeft?: boolean;
  topRight?: boolean;
  topLeft?: boolean;
}

export interface Instruction {
  ID: string;
  X: number;
  Y: number;
  GridAssetId: string;
  ClockwiseRotations: number;
}

export interface Blueprint {
  Tiles: TileData[][];
  Instructions: Instruction[];
  Ground?: Cell[][];
  DefaultTileColor: string;
  DefaultTileColor1: string;
}

export interface AreaDescription {
  Name: string;
  Safe: boolean;
  Blueprint: Blueprint;
  Transports: Transport[];
  North?: string;
  South?: string;
  East?: string;
  West?: string;
  Weather?: string;
  LoadStrategy?: string;
  SpawnStrategy?: string;
  BroadcastGroup?: string;
}

export interface Space {
  CollectionName: string;
  Name: string;
  Topology: string;
  Latitude: number;
  Longitude: number;
  AreaHeight: number;
  AreaWidth: number;
  Areas: AreaDescription[];
}

export interface Transport {
  SourceY: number;
  SourceX: number;
  DestY: number;
  DestX: number;
  DestStage: string;
  Confirmation?: boolean;
  RejectInteractable?: boolean;
}

export interface Prototype {
  id: string;
  commonName: string;
  cssColor: string;
  walkable: boolean;
  layer1css: string;
  layer2css: string;
  ceiling1css: string;
  ceiling2css: string;
  setName: string;
  mapColor: string;
  editorColor: string;
  displayText: string;
}

export interface Fragment {
  id: string;
  name: string;
  setName: string;
  blueprint: Blueprint;
}

export interface ReactionRule {
  reactsWith: string;
  reactsWithArgs: string[];
  reaction: string;
  reactionArgs: string[];
}

export interface InteractableStateDescription {
  cssClass: string;
  pushable: boolean;
  walkable: boolean;
  fragile: boolean;
  rejectTeleport?: boolean;
  reactions: string;
  reactionRules: ReactionRule[];
}

export interface InteractableDescription {
  id: string;
  name: string;
  setName: string;
  state: string;
  defaultState?: string;
  states?: Record<string, InteractableStateDescription>;
  cssClass: string;
  pushable: boolean;
  walkable: boolean;
  fragile: boolean;
  rejectTeleport?: boolean;
  reactions: string;
  reactionRules: ReactionRule[];
}

export interface Collection {
  Name: string;
  Spaces: Record<string, Space>;
  Fragments: Record<string, Fragment[]>;
  PrototypeSets: Record<string, Prototype[]>;
  InteractableSets: Record<string, InteractableDescription[]>;
}

export interface BootstrapResponse {
  collections: Record<string, Collection>;
  colors: Color[];
}

export interface Material {
  walkable?: boolean;
  ground1css?: string;
  ground2css?: string;
  layer1css?: string;
  layer2css?: string;
  ceiling1css?: string;
  ceiling2css?: string;
  displayText?: string;
}

export interface GridSelection {
  y: number;
  x: number;
}
