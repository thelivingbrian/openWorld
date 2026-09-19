import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { EditorApiService } from './editor-api.service';

describe('Hosted editor API', () => {
  let api: EditorApiService;
  let http: HttpTestingController;
  beforeEach(() => {
    history.replaceState({}, '', '/design/?world=test-world');
    TestBed.configureTestingModule({ providers: [provideHttpClient(), provideHttpClientTesting()] });
    api = TestBed.inject(EditorApiService); http = TestBed.inject(HttpTestingController);
  });
  afterEach(() => { http.verify(); history.replaceState({}, '', '/'); });

  async function bootstrap(): Promise<void> {
    const pending = api.getBootstrap();
    http.expectOne('/api/design/worlds/test-world/draft').flush({
      world: { id: 'test-world', name: 'World' }, resources: [
        { kind: 'collection', key: 'source', revision: 1, content: { Name: 'world', Spaces: {}, PrototypeSets: {}, InteractableSets: {}, Fragments: {} } },
        { kind: 'space', key: 'rooms', revision: 4, content: { Name: 'rooms', Areas: [] } },
        { kind: 'manifest', key: 'world', revision: 2, content: { name: 'World' } },
      ],
    });
    const data = await pending;
    expect(data.collections['world'].Spaces['rooms'].Areas).toEqual([]);
    expect(data.collections['world'].Manifest?.name).toBe('World');
  }

  it('preserves revisions when reading status and sends CSRF plus If-Match for saves', async () => {
    await bootstrap();
    const status = api.getWorldState();
    http.expectOne('/api/design/worlds/test-world').flush({ world: { id: 'test-world' } }); await status;
    const pending = api.saveSpace('world', 'rooms', { Name: 'rooms', Areas: [] } as any);
    http.expectOne('/api/csrf').flush({ token: 'csrf-token' });
    await Promise.resolve(); await Promise.resolve();
    const save = http.expectOne('/api/design/worlds/test-world/resources/space/rooms');
    expect(save.request.headers.get('If-Match')).toBe('4');
    expect(save.request.headers.get('X-CSRF-Token')).toBe('csrf-token');
    save.flush({ revision: 5 }); await pending;
  });

  it('creates connected spaces with weather and broadcast settings', async () => {
    await bootstrap();
    const pending = api.createSpace({ collectionName: 'world', name: 'garden', topology: 'torus', latitude: 2, longitude: 2, areaWidth: 2, areaHeight: 2, tileColor: 'green', tileColor1: 'brown', weather: 'raining', broadcastGroup: 'garden' });
    http.expectOne('/api/csrf').flush({ token: 'csrf-token' });
    await Promise.resolve(); await Promise.resolve();
    const save = http.expectOne('/api/design/worlds/test-world/resources/space/garden');
    const first = save.request.body.Areas[0];
    expect(first.North).toBe('garden:1-0'); expect(first.East).toBe('garden:0-1');
    expect(first.Weather).toBe('raining'); expect(first.BroadcastGroup).toBe('garden');
    save.flush({ revision: 1 }); await pending;
  });
});
