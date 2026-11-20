#!/usr/bin/env python3
"""Rebuild the editable Willow Reach seed assets. No running world is changed.

Run from any directory: python3 tools/worlds/build_willow_reach.py
The authored JSON files remain ordinary editor inputs; rerunning replaces only
this generator's Willow Reach assets. Commit editor changes before regenerating.
"""
from collections import deque
import json
import math
from pathlib import Path
import random

ROOT = Path(__file__).resolve().parents[2]
COLLECTION = ROOT / 'tools/main/data/collections/bloop'
SET = 'willow-reach'
N = 48
SPAWN = (23, 17)
COTTAGES = [(19, 13), (13, 23), (35, 39)]
COLORS = {
    'willow-grass': '#91b77e', 'willow-moss': '#739d72',
    'willow-path': '#dfc995', 'willow-water': '#4e939c',
    'willow-deep': '#357986', 'willow-wood': '#b68d61',
    'willow-stone': '#a8b3a4', 'willow-snow': '#dce7d5',
    'willow-shadow': '#29493f',
}
prototypes = []


def proto(name, *, color='', walkable=True, floor='', scenery='', text='', map_color=''):
    value = {
        'id': f'willow-{name}', 'commonName': name.replace('-', ' ').title(),
        'setName': SET, 'cssColor': color, 'walkable': walkable,
        'layer1css': floor, 'layer2css': scenery,
        'ceiling1css': '', 'ceiling2css': '',
        'mapColor': map_color or color or 'willow-grass', 'editorColor': '',
        'displayText': text,
    }
    prototypes.append(value)
    return value['id']


def tile(name):
    return {'prototypeId': f'willow-{name}'}


def blueprint(tiles, ground=None, primary='willow-grass', secondary='willow-path'):
    result = {'tiles': tiles, 'Instructions': [],
              'DefaultTileColor': primary, 'DefaultTileColor1': secondary}
    if ground is not None:
        result['Ground'] = ground
    return result


def area(name, bp, transports=None):
    return {'name': name, 'safe': True, 'blueprint': bp,
            'transports': transports or [], 'loadStrategy': '',
            'spawnStrategy': 'none', 'mapId': '', 'weather': '',
            'broadcastGroup': 'willow-reach'}


def transport(y, x, stage, dest_y, dest_x, confirmation=False):
    return {'sourceY': y, 'sourceX': x, 'destStage': stage,
            'destY': dest_y, 'destX': dest_x,
            'confirmation': confirmation, 'rejectInteractable': True}


def space(name, areas, size, topology='plane'):
    return {'CollectionName': 'bloop', 'Name': name, 'Topology': topology,
            'Latitude': 1, 'Longitude': len(areas),
            'AreaHeight': size, 'AreaWidth': size, 'Areas': areas}


def write(relative, content):
    path = COLLECTION / relative
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(content, indent=2) + '\n')


proto('ground')
proto('moss', color='willow-moss')
proto('snow', color='willow-snow')
proto('boundary', color='willow-moss', walkable=False)
proto('footprint', walkable=False)
proto('bridge', color='willow-wood', floor='willow-path-b thin no-lr')
proto('bridge-rail', color='willow-wood', walkable=False, floor='chocolate-b thick no-lr')
proto('stone', color='willow-stone')
proto('wood-floor', color='willow-wood')
proto('interior-wall', color='willow-path', walkable=False)
proto('interior-void', color='willow-shadow', walkable=False)
proto('rug', color='willow-deep')
proto('threshold', color='willow-path', floor='willow-wood s-hoz', text='Outside, the river is still singing.')
for name in ['oak', 'pine', 'birch', 'rock', 'bench', 'lantern']:
    proto(name, scenery=f'scenery-{name}', walkable=False,
          map_color='willow-moss' if name in ['oak','pine','birch'] else 'willow-wood')
proto('flower', scenery='scenery-flower')
proto('well', scenery='scenery-well', walkable=False, map_color='willow-stone')
proto('cottage', scenery='scenery-cottage', text='A little room, a little rest. Welcome in.', map_color='willow-wood')
proto('arrival', text='WILLOW REACH · Follow the lanterns. Find a little wonder.')
proto('bridge-sign', scenery='scenery-sign', text='RIVERWALK · Across the bridge: Pinewatch and the snow grove.')
proto('village-sign', scenery='scenery-sign', text='WILLOW GREEN · Cottages north. Riverwalk east. Old Bloopworld south.')
proto('pine-sign', scenery='scenery-sign', text='PINEWATCH · The first snow settles here. The village is across the river.')
proto('play-sign', scenery='scenery-sign', text='THE SEED GARDEN · Nudge the golden seed into the little stone ring.')
proto('legacy-sign', scenery='scenery-sign', text='THE OLD ROAD · Step into the arch to visit the original Bloopworld.')
proto('portal', color='willow-stone', floor='willow-path-b thick r1', scenery='sparkle', text='THE OLD ROAD · Continue to the original Bloopworld?')
proto('return-portal', color='willow-stone', floor='willow-path-b thick r1', text='WILLOW REACH · Take the woodland road back to the village.')
proto('goal', floor='willow-stone-b thick r1')
proto('welcome-room', color='willow-wood', text='THE KEEPER’S COTTAGE · Leave by the south door to return to Willow Reach.')
for mask in range(16):
    corners = ' '.join(f'r0-{corner}' for bit, corner in enumerate(['tl','tr','bl','br']) if mask & (1<<bit))
    proto(f'water-{mask}', color='willow-grass', floor=f'willow-water {corners}'.strip(), walkable=False, map_color='willow-water')
proto('water-glint', color='willow-water', scenery='sparkle', walkable=False)

rng = random.Random(7419)
tiles = [[tile('ground') for _ in range(N)] for _ in range(N)]
paths = set()
water = set()
reserved = set()
for y in range(N):
    center = 31 + round(2 * math.sin(y / 7))
    for x in range(center-2, center+3):
        water.add((y,x))
    for x in range(N):
        if y in (0,N-1) or x in (0,N-1): tiles[y][x] = tile('boundary')
        elif (math.sin(x/5) + math.cos(y/4) > 1.1 and x < center-4):
            tiles[y][x] = tile('moss')
        elif x > 35 and y < 17:
            tiles[y][x] = tile('snow')


def line(points, radius=1):
    for (ay,ax),(by,bx) in zip(points,points[1:]):
        steps = max(abs(by-ay),abs(bx-ax),1)
        for step in range(steps+1):
            y = round(ay+(by-ay)*step/steps)
            x = round(ax+(bx-ax)*step/steps)
            for dy in range(-radius,radius+1):
                for dx in range(-radius,radius+1):
                    yy,xx=y+dy,x+dx
                    if 1 <= yy < N-1 and 1 <= xx < N-1:
                        paths.add((yy,xx))
                        tiles[yy][xx]=tile('ground')


# A village loop, a river crossing, and two quieter woodland detours.
line([(44,17),(33,17),(29,17),(25,17),(22,14),(20,13)])
line([(25,17),(25,25),(25,38),(30,39),(36,39)])
line([(23,14),(19,20),(14,23)])
line([(25,38),(20,39),(12,41),(7,39)])
line([(29,17),(32,11),(36,11),(38,18),(33,24),(25,24)])
line([(23,14),(23,8),(28,7),(32,11)])
for y in range(23,31):
    for x in range(13,22):
        if ((y-26.5)/4.6)**2+((x-17)/5.4)**2 < 1:
            paths.add((y,x)); tiles[y][x]=tile('ground')

# Smooth the river's outside corners against meadow, preserving blocked water.
for y,x in water:
    mask=0
    for bit,(dy,dx) in enumerate([(-1,-1),(-1,1),(1,-1),(1,1)]):
        if (y+dy,x) not in water and (y,x+dx) not in water: mask |= 1<<bit
    tiles[y][x]=tile(f'water-{mask}')
    if mask==0 and rng.random()<0.07: tiles[y][x]=tile('water-glint')
# Three-tile-wide bridge: continuous walkable boards and solid outer rails.
for y in range(23,28):
    for x in range(26,37):
        tiles[y][x]=tile('bridge-rail' if y in (23,27) else 'bridge')
        reserved.add((y,x))

for y,x in COTTAGES:
    for yy in range(y-4,y+3):
        for xx in range(x-3,x+4): reserved.add((yy,xx))
    for yy in range(y-1,y+1):
        for xx in range(x-2,x+3): tiles[yy][xx]=tile('footprint')
    tiles[y][x]=tile('cottage')
    # Keep a warm, visible threshold outside each door.
    paths.add((y+1,x)); tiles[y+1][x]=tile('ground')

# Broad sightlines around paths; forest clusters frame, rather than fill, routes.
trees=[]
for y in range(3,N-2):
    for x in range(3,N-2):
        if (y,x) in water or (y,x) in reserved: continue
        if any((y+dy,x+dx) in paths for dy in range(-2,3) for dx in range(-2,3)): continue
        if any(abs(y-ty)<3 and abs(x-tx)<3 for ty,tx in trees): continue
        if rng.random() > .22: continue
        kind='pine' if x>35 else ('birch' if y>31 and x<14 else 'oak')
        trees.append((y,x)); tiles[y][x]=tile(kind)

# Designed details, all outside the primary travel line.
details = {
    (28,14):'lantern', (28,20):'lantern', (22,25):'lantern', (28,37):'lantern',
    (20,10):'lantern', (20,16):'flower', (14,20):'flower', (36,36):'lantern',
    (26,12):'bench', (30,21):'bench', (21,38):'bench',
    (30,15):'village-sign', (24,24):'bridge-sign', (18,40):'pine-sign',
    (36,13):'play-sign', (43,18):'legacy-sign',
    SPAWN:'arrival', (44,17):'portal', (35,11):'goal',
}
for (y,x),name in details.items():
    tiles[y][x]=tile(name)
# A low village fountain creates a focal point without hiding the paths.
for yy in (27,28):
    for xx in (16,17,18): tiles[yy][xx]=tile('footprint')
tiles[28][17]=tile('well')
# A pushable garden seed uses existing interaction mechanics.
tiles[37][11]['interactableId']='willow-seed'
# Small wildflower drifts and occasional river stones.
for y,x in [(21,7),(22,6),(23,6),(31,24),(32,25),(33,25),(12,17),(13,17),
            (17,5),(18,6),(19,5),(37,20),(38,21),(39,21),(10,38),(9,39),
            (30,42),(31,43),(32,43),(40,7),(41,8)]:
    if (y,x) not in water and tiles[y][x]['prototypeId'] in ['willow-ground','willow-moss','willow-snow']:
        tiles[y][x]=tile('flower')
for y,x in [(18,27),(31,29),(38,28),(9,29),(39,35),(7,37),(42,12)]:
    if (y,x) not in water and (y,x) not in paths: tiles[y][x]=tile('rock')

# Ground paint gives paths rounded convex corners without affecting footprints.
ground=[]
for y in range(N):
    row=[]
    for x in range(N):
        status=int((y,x) in paths)
        cell={'Status':status}
        for name,dy,dx in [('TopLeft',-1,-1),('TopRight',-1,1),('BottomLeft',1,-1),('BottomRight',1,1)]:
            if int((y+dy,x) in paths)!=status and int((y,x+dx) in paths)!=status:
                cell[name]=True
        row.append(cell)
    ground.append(row)

transports=[transport(y,x,f'willow-cottage:0-{i}',12,8) for i,(y,x) in enumerate(COTTAGES)]
transports.append(transport(44,17,'infirmary:0-0',9,9,True))
region=space(SET,[area('willow-reach:0-0',blueprint(tiles,ground),transports)],N)

interiors=[]
for i,(door_y,door_x) in enumerate(COTTAGES):
    room=[[tile('interior-void') for _ in range(16)] for _ in range(16)]
    for y in range(2,14):
        for x in range(2,14):
            room[y][x]=tile('interior-wall' if y in (2,13) or x in (2,13) else 'wood-floor')
    for y in range(7,11):
        for x in range(6,11): room[y][x]=tile('rug')
    for y,x in [(4,4),(4,11),(12,5)]: room[y][x]=tile('lantern')
    for y,x in [(5,4),(5,10)]: room[y][x]=tile('bench')
    room[4][8]=tile('flower')
    room[12][8]=tile('welcome-room')
    room[13][8]=tile('threshold')
    interiors.append(area(f'willow-cottage:0-{i}',blueprint(room,primary='willow-wood'),
                          [transport(13,8,'willow-reach:0-0',door_y+1,door_x)]))

fragments=[]
for name in ['oak','pine','birch','lantern','sign','rock','flower','bench']:
    source='village-sign' if name=='sign' else name
    fragments.append({'id':f'willow-fragment-{name}','name':f'Willow {name.title()}',
                      'setName':SET,'blueprint':blueprint([[tile(source)]])})
well_tiles=[[tile('footprint') for _ in range(3)] for _ in range(2)]
well_tiles[1][1]=tile('well')
fragments.append({'id':'willow-fragment-well','name':'Willow Village Fountain', 'setName':SET,'blueprint':blueprint(well_tiles)})
cottage_tiles=[[tile('footprint') for _ in range(5)] for _ in range(2)]
cottage_tiles[1][2]=tile('cottage')
fragments.append({'id':'willow-fragment-cottage','name':'Willow Cottage · place transport at bottom center',
                  'setName':SET,'blueprint':blueprint(cottage_tiles)})
bridge_tiles=[[tile('bridge-rail' if y in (0,4) else 'bridge') for _ in range(11)] for y in range(5)]
fragments.append({'id':'willow-fragment-bridge','name':'Willow River Bridge',
                  'setName':SET,'blueprint':blueprint(bridge_tiles)})

# Verify landmarks and every transport source are reachable from the arrival.
by_id={p['id']:p for p in prototypes}
seen={SPAWN}; pending=deque([SPAWN])
while pending:
    y,x=pending.popleft()
    for yy,xx in [(y-1,x),(y+1,x),(y,x-1),(y,x+1)]:
        if 0<=yy<N and 0<=xx<N and (yy,xx) not in seen and by_id[tiles[yy][xx]['prototypeId']]['walkable']:
            seen.add((yy,xx)); pending.append((yy,xx))
for point in [*COTTAGES,(44,17),(25,38),(7,39),(35,11),(37,11)]:
    assert point in seen, f'Unreachable landmark: {point}'
assert by_id[tiles[SPAWN[0]][SPAWN[1]]['prototypeId']]['walkable']
for i,room in enumerate(interiors):
    bp=room['blueprint']['tiles']
    assert by_id[bp[12][8]['prototypeId']]['walkable']
    assert by_id[bp[13][8]['prototypeId']]['walkable']
# Check the connection into existing Bloopworld against source data.
legacy_space=json.loads((COLLECTION/'spaces/infirmary.json').read_text())
legacy=next(a for a in legacy_space['Areas'] if a['name']=='infirmary:0-0')
legacy_id=legacy['blueprint']['tiles'][9][9].get('prototypeId','')
legacy_prototypes=[p for f in (COLLECTION/'prototypes').glob('*.json') for p in json.loads(f.read_text())]
assert next(p for p in legacy_prototypes if p['id']==legacy_id)['walkable']

# A distinct arrival and departure tile keeps the old-world connection reversible.
legacy['blueprint']['tiles'][9][10]=tile('return-portal')
legacy['transports']=[t for t in legacy['transports'] if not (t['sourceY']==9 and t['sourceX']==10)]
legacy['transports'].append(transport(9,10,'willow-reach:0-0',43,17,True))
write('spaces/infirmary.json',legacy_space)

write('prototypes/willow-reach.json',prototypes)
write('fragments/willow-reach.json',fragments)
write('spaces/willow-reach.json',region)
write('spaces/willow-cottage.json',space('willow-cottage',interiors,16,'disconnected'))
write('interactables/willow-reach.json',[{
    'id':'willow-seed','name':'Garden seed','setName':SET,
    'cssClass':'gold r1 willow-wood-b thin','pushable':True,
    'walkable':False,'fragile':False,'reactions':'',
}])
palette_path=ROOT/'tools/main/data/colors/colors.json'
palette=json.loads(palette_path.read_text())
for name,color in COLORS.items():
    if not any(c['cssClassName']==name for c in palette):
        palette.append({'cssClassName':name,'R':int(color[1:3],16),'G':int(color[3:5],16),'B':int(color[5:7],16),'A':''})
palette_path.write_text(json.dumps(palette,indent=2)+'\n')
print(f'Willow Reach: {N}×{N}, {len(trees)} trees, {len(seen)} reachable tiles, {len(prototypes)} prototypes, {len(fragments)} fragments.')
print(f'Arrival y,x={SPAWN}; cottage doors={COTTAGES}; old-world portal=(44,17).')
