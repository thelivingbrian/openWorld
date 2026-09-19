export type Lifecycle = 'owner-present' | 'until-empty' | 'persistent';
export interface WorldLocation { stage: string; y: number; x: number; }
export interface WorldTeam { id: string; label: string; color: string; spawn: WorldLocation; }
export interface NPCStep {
  action: 'wait' | 'north' | 'south' | 'east' | 'west' | 'wander' | 'chase' | 'attack';
  ticks: number;
  condition: 'always' | 'hurt' | 'player-nearby';
  radius: number;
  damage: number;
}
export interface NPCDefinition {
  id: string; name: string; team: string; appearance: string; health: number;
  money: number; intervalMs: number; lifetimeSeconds: number; program: NPCStep[];
}
export interface SpawnRule {
  id: string; stage: string; kind: 'npc' | 'money' | 'boost' | 'powerup'; npc: string;
  trigger: 'once' | 'entry'; count: number; maxAlive: number; cooldownSeconds: number;
  randomPosition: boolean; y: number; x: number; amount: number;
}
export interface WorldAchievement {
  id: string; name: string; description: string; metric: string; target: number; stage: string;
}
export interface WorldManifest {
  name: string; description?: string; entry: WorldLocation; teams: WorldTeam[];
  defaultTeam: string; maxPlayers: number; lifecycle: Lifecycle;
  leaderboards?: { id: string; label: string; metric: string }[];
  onboardingStages?: string[]; onboardingExit?: WorldLocation;
  npcs: NPCDefinition[]; spawns: SpawnRule[]; achievements: WorldAchievement[];
}
export interface WorldDocument {
  id: string; name: string; slug?: string; visibility: 'public' | 'unlisted';
  moderationState: string; lifecycle: Lifecycle; deploymentEnabled: boolean;
  draftGeneration: number; publishedReleaseId?: string;
}
export interface WorldRelease {
  id: string; number: number; label?: string; createdAt: string; draftGeneration: number;
}
export interface RuntimeInfo { releaseId: string; state: string; playerCount: number; }

export function defaultManifest(name: string, stage: string): WorldManifest {
  return { name, entry: { stage, y: 0, x: 0 }, teams: [], defaultTeam: '', maxPlayers: 100,
    lifecycle: 'persistent', npcs: [], spawns: [], achievements: [] };
}
