import { WorldSettingsComponent } from './world-settings.component';
import { defaultManifest } from '../../core/models/world.models';

describe('world settings for recurring population and relays', () => {
  it('creates bounded recurring rules and preserves valid puzzle points after malformed input', () => {
    const component = new WorldSettingsComponent();
    component.manifest = defaultManifest('Frontier', 'room');
    component.stages = ['room'];
    component.addSpawn();
    expect(component.manifest.spawns[0]).toMatchObject({trigger:'interval',maxAlive:5,cooldownSeconds:30});
    component.toggleRegion(component.manifest.spawns[0]);
    expect(component.manifest.spawns[0].region).toEqual({x:0,y:0,width:8,height:8});
    component.addPuzzle();
    const puzzle = component.manifest.puzzles![0];
    component.setPoints(puzzle,'plates','9,5; 9,11');
    expect(puzzle.plates).toEqual([{y:9,x:5},{y:9,x:11}]);
    component.setPoints(puzzle,'plates','not a coordinate');
    expect(puzzle.plates).toEqual([{y:9,x:5},{y:9,x:11}]);
    expect(component.pointError).toContain('Y,X');
  });
});
