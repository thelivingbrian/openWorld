/**
 * Registry of available ReactsWith gates and Reaction actions
 * for composing interactable behavior from the design workspace.
 *
 * Each entry describes the function name (matching the server-side registry key),
 * a human-readable label, and the arguments it expects.
 */

export interface RegistryArg {
  /** Argument name shown in the editor */
  name: string;
  /** Hint text or placeholder */
  placeholder: string;
}

export interface RegistryEntry {
  /** Key that maps to the server-side registry */
  key: string;
  /** Human-readable label */
  label: string;
  /** Ordered list of expected arguments */
  args: RegistryArg[];
}

// ── ReactsWith gates ──────────────────────────────────────────────────────

export const REACTS_WITH_REGISTRY: RegistryEntry[] = [
  { key: 'everything',           label: 'Everything (always matches)',               args: [] },
  { key: 'never',                label: 'Never (never matches)',                     args: [] },
  { key: 'interactableIsNil',    label: 'No incoming interactable (player only)',     args: [] },
  { key: 'interactableIsABall',  label: 'Incoming is a ball',                        args: [] },
  { key: 'interactableIsARing',  label: 'Incoming is a ring',                        args: [] },
  { key: 'interactableHasName',  label: 'Incoming has specific name',                args: [{ name: 'name', placeholder: 'e.g. ball-fuchsia' }] },
  { key: 'interactableStateIs', label: 'Incoming state is', args: [{ name: 'state', placeholder: 'e.g. armed' }] },
  { key: 'interactableStateIsNot', label: 'Incoming state is not', args: [{ name: 'state', placeholder: 'e.g. disarmed' }] },
  { key: 'interactableStateContains', label: 'Incoming state contains', args: [{ name: 'fragment', placeholder: 'e.g. open' }] },
  { key: 'playerHasTeam',        label: 'Player belongs to team',                    args: [{ name: 'team', placeholder: 'e.g. sky-blue' }] },
  { key: 'playerTeamAndBallNameMatch', label: 'Player team matches ball name',       args: [{ name: 'team', placeholder: 'e.g. fuchsia' }] },
  { key: 'PlayerAndTeamMatchButDifferentBall', label: 'Team matches but different ball', args: [{ name: 'team', placeholder: 'e.g. sky-blue' }] },
];

// ── Reaction actions ──────────────────────────────────────────────────────

export const REACTION_REGISTRY: RegistryEntry[] = [
  { key: 'eat',                  label: 'Eat (discard incoming)',                     args: [] },
  { key: 'pass',                 label: 'Pass (push through)',                        args: [] },
  { key: 'transmitPushAll',      label: 'Transmit push to all interactables',         args: [] },
  { key: 'transmitPushByState',  label: 'Transmit push by interactable state',        args: [{ name: 'state', placeholder: 'e.g. armed' }] },
  { key: 'transmitPushByName',   label: 'Transmit push by interactable name',         args: [{ name: 'name', placeholder: 'e.g. box' }] },
  { key: 'killInstantly',        label: 'Kill player instantly',                      args: [] },
  { key: 'playSoundForAll',      label: 'Play sound for all nearby',                  args: [{ name: 'soundName', placeholder: 'e.g. explosion' }] },
  { key: 'playSoundForInitiator', label: 'Play sound for initiator',                  args: [{ name: 'soundName', placeholder: 'e.g. water-splash' }] },
  { key: 'notifyAndPass',        label: 'Notify player and pass',                     args: [{ name: 'message', placeholder: 'notification text' }] },
  { key: 'hideByTeam',           label: 'Hide on enemy team area',                    args: [{ name: 'team', placeholder: 'e.g. fuchsia' }] },
  { key: 'scoreGoalForTeam',     label: 'Score goal for team',                        args: [{ name: 'team', placeholder: 'e.g. sky-blue' }] },
  { key: 'showScoreToPlayer',    label: 'Show score to player',                       args: [{ name: 'team', placeholder: 'e.g. fuchsia' }] },
  { key: 'catapultWest',         label: 'Catapult west',                              args: [] },
  { key: 'catapultEast',         label: 'Catapult east',                              args: [] },
  { key: 'catapultNorth',        label: 'Catapult north',                             args: [] },
  { key: 'catapultSouth',        label: 'Catapult south',                             args: [] },
  { key: 'moveInitiator',        label: 'Move player by offset',                     args: [{ name: 'yOff', placeholder: '0' }, { name: 'xOff', placeholder: '0' }] },
  { key: 'destroyInRangeSkipingSelf', label: 'Destroy in range (skip self)',          args: [
    { name: 'yMin', placeholder: '0' },
    { name: 'xMin', placeholder: '0' },
    { name: 'yMax', placeholder: '0' },
    { name: 'xMax', placeholder: '0' },
  ]},
  { key: 'makeDangerousForOtherTeam', label: 'Make dangerous for other team',        args: [] },
  { key: 'damageAndSpawn',       label: 'Damage and spawn',                           args: [] },
  { key: 'teleportHomeInteraction', label: 'Teleport home interaction',               args: [] },
];

/** Lookup helpers */
export function findReactsWithEntry(key: string): RegistryEntry | undefined {
  return REACTS_WITH_REGISTRY.find(e => e.key === key);
}

export function findReactionEntry(key: string): RegistryEntry | undefined {
  return REACTION_REGISTRY.find(e => e.key === key);
}
