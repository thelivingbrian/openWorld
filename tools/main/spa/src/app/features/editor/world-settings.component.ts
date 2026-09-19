import { Component, EventEmitter, Input, Output } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NPCDefinition, NPCStep, WorldManifest } from '../../core/models/world.models';

@Component({
  selector: 'app-world-settings',
  imports: [FormsModule],
  templateUrl: './world-settings.component.html',
  styleUrl: './world-settings.component.css',
})
export class WorldSettingsComponent {
  @Input({ required: true }) manifest!: WorldManifest;
  @Input() mode = 'settings';
  @Input() stages: string[] = [];
  @Output() changed = new EventEmitter<void>();

  readonly actions = ['wait', 'north', 'south', 'east', 'west', 'wander', 'chase', 'attack'];
  readonly metrics = [
    { id: 'visit-stage', label: 'Visit a stage' }, { id: 'peakWealth', label: 'Peak money' },
    { id: 'peakKillStreak', label: 'Peak kill streak' }, { id: 'killCount', label: 'Player defeats' },
    { id: 'killCountNpc', label: 'NPC defeats' }, { id: 'goalsScored', label: 'Goals scored' },
    { id: 'deathCount', label: 'Deaths' },
  ];

  private nextID(prefix: string, values: { id: string }[]): string {
    let index = 1;
    while (values.some(value => value.id === `${prefix}-${index}`)) index++;
    return `${prefix}-${index}`;
  }

  addNPC(): void {
    this.manifest.npcs.push({ id: this.nextID('npc', this.manifest.npcs), name: 'New NPC', team: 'npc',
      appearance: 'red-b thick r0', health: 100, money: 20, intervalMs: 400, lifetimeSeconds: 300,
      program: [{ action: 'wander', ticks: 4, condition: 'always', radius: 4, damage: 25 }] });
    this.changed.emit();
  }

  addStep(npc: NPCDefinition): void {
    npc.program.push({ action: 'wait', ticks: 1, condition: 'always', radius: 4, damage: 25 });
    this.changed.emit();
  }

  moveStep(steps: NPCStep[], index: number, offset: number): void {
    const destination = index + offset;
    if (destination < 0 || destination >= steps.length) return;
    [steps[index], steps[destination]] = [steps[destination], steps[index]];
    this.changed.emit();
  }

  addSpawn(): void {
    this.manifest.spawns.push({ id: this.nextID('spawn', this.manifest.spawns), stage: this.stages[0] ?? '',
      kind: this.manifest.npcs.length ? 'npc' : 'money', npc: this.manifest.npcs[0]?.id ?? '',
      trigger: 'entry', count: 1, maxAlive: 5, cooldownSeconds: 30, randomPosition: true, y: 0, x: 0, amount: 10 });
    this.changed.emit();
  }

  addAchievement(): void {
    this.manifest.achievements.push({ id: this.nextID('achievement', this.manifest.achievements),
      name: 'New achievement', description: '', metric: 'visit-stage', target: 1, stage: this.stages[0] ?? '' });
    this.changed.emit();
  }

  addTeam(): void {
    const id = this.nextID('team', this.manifest.teams);
    this.manifest.teams.push({ id, label: 'New team', color: 'sky-blue', spawn: { ...this.manifest.entry } });
    if (!this.manifest.defaultTeam) this.manifest.defaultTeam = id;
    this.changed.emit();
  }

  remove<T>(items: T[], index: number): void {
    items.splice(index, 1);
    this.changed.emit();
  }
}
