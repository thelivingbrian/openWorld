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
    'annex-void': '#111c2b', 'annex-floor': '#293746',
    'annex-wall': '#46515f', 'annex-edge': '#718293',
    'annex-cyan': '#8bd9dc', 'annex-lilac': '#b7a4df',
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


def area(name, bp, transports=None, safe=True):
    return {'name': name, 'safe': safe, 'blueprint': bp,
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
proto('village-sign', scenery='scenery-sign', text='WILLOW GREEN · Two crossings, roaming danger. The quiet puzzle annex lies south.')
proto('pine-sign', scenery='scenery-sign', text='PINEWATCH · The first snow settles here. The village is across the river.')
proto('play-sign', scenery='scenery-sign', text='THE SEED GARDEN · Nudge the golden seed into the little stone ring.')
proto('annex-sign', scenery='scenery-sign', text='THE QUIET ANNEX · Optional puzzles for one or two. No fighting. No health reset.')
proto('annex-portal', color='annex-floor', floor='annex-lilac-b thick r0', text='THE QUIET ANNEX · Step inside to explore the solo and two-player chambers.')
proto('sky-refuge', color='willow-stone', floor='sky-blue-b thick r0', text='SKY BLUE REFUGE · A quiet place. Entering never restores health.')
proto('pink-refuge', color='willow-stone', floor='fuchsia-b thick r0', text='FUCHSIA REFUGE · A quiet place. Entering never restores health.')
proto('refuge-exit', color='willow-path', floor='willow-stone-b thick r0', text='THE FRONTIER · Power uses your collected shape. Toggle Boost for longer steps. Health only returns after death.')
proto('refuge-welcome', color='willow-path', text='YOUR REFUGE · Collect the power and boost ahead, then follow the path north. Fight the orange wild bloops outside.')
proto('goal', floor='willow-stone-b thick r1')
proto('welcome-room', color='willow-wood', text='THE KEEPER’S COTTAGE · Leave by the south door to return to Willow Reach.')
for mask in range(16):
    corners = ' '.join(f'r0-{corner}' for bit, corner in enumerate(['tl','tr','bl','br']) if mask & (1<<bit))
    proto(f'water-{mask}', color='willow-grass', floor=f'willow-water {corners}'.strip(), walkable=False, map_color='willow-water')
proto('water-glint', color='willow-water', scenery='sparkle', walkable=False)
proto('snowpine', color='willow-snow', scenery='scenery-snowpine', walkable=False, map_color='willow-snow')
proto('snow-rock', color='willow-snow', scenery='scenery-rock', walkable=False)
proto('snow-flower', color='willow-snow', scenery='scenery-flower')
proto('annex-void', color='annex-void', walkable=False)
proto('annex-floor', color='annex-floor')
proto('annex-wall', color='annex-wall', floor='annex-edge-b thin no-lr', walkable=False)
proto('annex-lamp', color='annex-wall', floor='annex-cyan-b thin r0', walkable=False)
proto('annex-plate', color='annex-floor', floor='annex-cyan-b med r1')
proto('annex-goal', color='annex-floor', floor='annex-lilac-b thick r0')
proto('annex-relay', color='annex-floor', floor='annex-lilac-b thick r1')
proto('annex-sky-exit', color='annex-floor', floor='sky-blue-b thick r0', text='SKY BLUE REFUGE · Leave without healing or resetting your streak.')
proto('annex-pink-exit', color='annex-floor', floor='fuchsia-b thick r0', text='FUCHSIA REFUGE · Leave without healing or resetting your streak.')
proto('annex-gate', color='annex-wall', floor='annex-lilac-b thick no-tb', walkable=False)
proto('annex-exit', color='annex-floor', floor='annex-edge-b thick r0', text='LEAVE · You keep your current health and streak. There is no combat reward here.')
proto('solo-entrance', color='annex-floor', floor='annex-cyan-b thick r0', text='ONE · Push the square block onto the round socket. Then cross the open gate to the goal.')
proto('duo-entrance', color='annex-floor', floor='annex-lilac-b thick r0', text='TWO · Two players, any teams. Stand together on the two sockets. Optional; the frontier stays open.')
proto('solo-hint', color='annex-floor', text='ONE · Push the block north onto the cyan socket. Leave freely. Menu → Reset chamber (everyone agrees).')
proto('duo-hint', color='annex-floor', text='TWO · One bloop on each socket, together. Blocks do not count. Leave freely. Menu → Reset chamber (everyone agrees).')
proto('annex-welcome', color='annex-floor', text='THE QUIET ANNEX · Left: one-player chamber. Right: two-player chamber. South: team refuges. No fighting or healing.')

def in_snow(y, x):
    return x >= 35 + round(math.sin(y / 4)) and y < 18 + round(2 * math.sin(x / 3))

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
        elif in_snow(y,x):
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
# The northern footbridge makes the river a route choice, not a single choke.
line([(19,20),(16,24),(10,26),(10,38),(12,41)])
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

for y in range(8,13):
    for x in range(28,38):
        tiles[y][x]=tile('bridge-rail' if y in (8,12) else 'bridge')
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
        kind=('snowpine' if in_snow(y,x) else 'pine') if x>35 else ('birch' if y>31 and x<14 else 'oak')
        trees.append((y,x)); tiles[y][x]=tile(kind)

# Designed details, all outside the primary travel line.
details = {
    (28,14):'lantern', (28,20):'lantern', (22,25):'lantern', (28,37):'lantern',
    (20,10):'lantern', (20,16):'flower', (14,20):'flower', (36,36):'lantern',
    (26,12):'bench', (30,21):'bench', (21,38):'bench',
    (30,15):'village-sign', (24,24):'bridge-sign', (18,40):'pine-sign',
    (36,13):'play-sign', (43,18):'annex-sign',
    SPAWN:'sky-refuge', (25,38):'pink-refuge', (44,17):'annex-portal', (35,11):'goal',
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
        tiles[y][x]=tile('snow-flower' if in_snow(y,x) else 'flower')
for y,x in [(18,27),(31,29),(38,28),(9,29),(39,35),(7,37),(42,12)]:
    if (y,x) not in water and (y,x) not in paths: tiles[y][x]=tile('snow-rock' if in_snow(y,x) else 'rock')

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
for mask in range(16):
    corners=' '.join(f'r0-{corner}' for bit,corner in enumerate(['tl','tr','bl','br']) if mask & (1<<bit))
    proto(f'snow-path-{mask}',color='willow-snow',floor=f'willow-path {corners}'.strip(),map_color='willow-path')
for y in range(1,N-1):
    for x in range(1,N-1):
        if not in_snow(y,x): continue
        if tiles[y][x]['prototypeId']=='willow-ground':
            if (y,x) in paths:
                mask=sum(1<<bit for bit,key in enumerate(['TopLeft','TopRight','BottomLeft','BottomRight']) if ground[y][x].get(key))
                tiles[y][x]=tile(f'snow-path-{mask}')
            else:
                tiles[y][x]=tile('snow')

# Rounded, irregular snow edges follow one coherent biome shape. Scenery keeps
# the same underlying snow rather than repainting a grass square at its anchor.
snow_variants={}
base_by_id={p['id']:p for p in prototypes}
for y in range(1,N-1):
    for x in range(1,N-1):
        if not in_snow(y,x) or (y,x) in water: continue
        base=base_by_id[tiles[y][x]['prototypeId']]
        if base['layer2css'].startswith('scenery-') and base['cssColor']!='willow-snow':
            name=base['id'].removeprefix('willow-')+'-snowbase'
            if name not in snow_variants:
                copy=dict(base);copy.update(id='willow-'+name,cssColor='willow-snow',commonName=base['commonName']+' · Snow')
                prototypes.append(copy);snow_variants[name]=copy
            tiles[y][x]=tile(name);base=snow_variants[name]
        if base['cssColor']=='willow-snow' and not base['layer1css']:
            mask=sum(1<<bit for bit,(dy,dx) in enumerate([(-1,-1),(-1,1),(1,-1),(1,1)]) if not in_snow(y+dy,x) or not in_snow(y,x+dx))
            if not mask: continue
            name=base['id'].removeprefix('willow-')+f'-edge-{mask}'
            if name not in snow_variants:
                corners=' '.join('r0-'+corner for bit,corner in enumerate(['tl','tr','bl','br']) if mask&(1<<bit))
                copy=dict(base);copy.update(id='willow-'+name,cssColor='willow-grass',layer1css='willow-snow '+corners)
                prototypes.append(copy);snow_variants[name]=copy
            tiles[y][x]=tile(name)

transports=[transport(y,x,f'willow-cottage:0-{i}',12,8) for i,(y,x) in enumerate(COTTAGES)]
transports.extend([transport(44,17,'willow-annex:0-0',12,8),
                   transport(23,17,'willow-refuge:0-0',4,6),
                   transport(25,38,'willow-refuge:0-1',4,6)])
region=space(SET,[area('willow-reach:0-0',blueprint(tiles,ground),transports,safe=False)],N)

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
for name in ['oak','pine','snowpine','birch','lantern','sign','rock','flower','bench']:
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

# Small protected arrival stages keep combat rules explicit at the stage boundary.
refuges=[]
for i,team in enumerate(['sky-blue','fuchsia']):
    refuge=[[tile('boundary') for _ in range(12)] for _ in range(12)]
    for y in range(2,10):
        for x in range(2,10): refuge[y][x]=tile('ground')
    for y in range(3,9):
        for x in range(5,8): refuge[y][x]=tile('stone')
    for y,x in [(3,3),(3,9)]: refuge[y][x]=tile('oak')
    for y,x in [(6,3),(8,9)]: refuge[y][x]=tile('flower')
    refuge[8][3]=tile('bench')
    refuge[7][6]=tile('refuge-welcome')
    refuge[3][6]=tile('refuge-exit')
    dest_y,dest_x=[(24,17),(26,38)][i]
    refuges.append(area(f'willow-refuge:0-{i}',blueprint(refuge),
                        [transport(3,6,'willow-reach:0-0',dest_y,dest_x)]))

# Low abstract geometry makes puzzle adjacency legible from above.
def dark_room():
    room=[[tile('annex-void') for _ in range(16)] for _ in range(16)]
    for y in range(2,14):
        for x in range(2,14):
            room[y][x]=tile('annex-wall' if y in (2,13) or x in (2,13) else 'annex-floor')
    for y,x in [(2,4),(2,11),(13,3),(13,12)]: room[y][x]=tile('annex-lamp')
    return room

foyer=dark_room()
foyer[5][4]=tile('solo-entrance');foyer[5][11]=tile('duo-entrance')
foyer[12][8]=tile('annex-welcome')
foyer[13][6]=tile('annex-sky-exit');foyer[13][9]=tile('annex-pink-exit')
foyer_area=area('willow-annex:0-0',blueprint(foyer,primary='annex-floor'),[
    transport(5,4,'willow-solo:0-0',12,8),
    transport(5,11,'willow-duo:0-0',12,8),
    transport(13,6,'willow-refuge:0-0',4,6),
    transport(13,9,'willow-refuge:0-1',4,6),
])
solo=dark_room()
for x in range(3,13): solo[6][x]=tile('annex-wall')
for x in [7,8,9]: solo[6][x]=tile('annex-gate')
solo[9][5]=tile('annex-plate')
solo[11][5]['interactableId']='willow-puzzle-block'
solo[4][8]=tile('annex-goal')
solo[12][8]=tile('solo-hint');solo[13][8]=tile('annex-exit')
solo_area=area('willow-solo:0-0',blueprint(solo,primary='annex-floor'),
               [transport(13,8,'willow-annex:0-0',6,4)])
duo=dark_room()
for x in range(3,13): duo[6][x]=tile('annex-wall')
for x in [7,8,9]: duo[6][x]=tile('annex-gate')
duo[9][5]=tile('annex-plate');duo[9][11]=tile('annex-relay')
duo[12][8]=tile('duo-hint');duo[13][8]=tile('annex-exit')
duo_area=area('willow-duo:0-0',blueprint(duo,primary='annex-floor'),
              [transport(13,8,'willow-annex:0-0',6,11)])

# Spawn rules are the single source of truth for supplies and encounters.
def spawn(id,stage,kind,y,x,*,interval=18,amount=1,**extra):
    return {'id':id,'stage':stage,'kind':kind,'trigger':'interval','count':1,
            'maxAlive':1,'cooldownSeconds':interval,'randomPosition':False,
            'y':y,'x':x,'amount':amount,'lifetimeSeconds':90,**extra}

spawns=[]
for i,team in enumerate(['sky-blue','fuchsia']):
    stage=f'willow-refuge:0-{i}'
    spawns.extend([
        spawn(f'{team}-starter-power',stage,'powerup',5,6,interval=15),
        spawn(f'{team}-starter-boost',stage,'boost',6,6,interval=12,amount=3),
    ])
for name,y,x in [('village',21,18),('pinewatch',28,39),('north-road',13,26),('garden',38,15)]:
    tiles[y][x]=tile('ground');tiles[y][x+1]=tile('ground')
    spawns.extend([
        spawn(f'{name}-power','willow-reach:0-0','powerup',y,x,interval=24),
        spawn(f'{name}-boost','willow-reach:0-0','boost',y,x+1,interval=18,amount=2),
    ])
for i,(name,y,x) in enumerate([('north-cache',9,25),('river-cache',20,40),('garden-cache',33,22)]):
    tiles[y][x]=tile('ground');tiles[y][x+1]=tile('ground')
    spawns.extend([
        spawn(name+'-power','willow-reach:0-0','powerup',y,x,interval=12,
              rotationGroup='frontier-supply',activeSeconds=60,label={'north-cache':'North Crossing','river-cache':'Riverbank','garden-cache':'Seed Garden'}[name]),
        spawn(name+'-boost','willow-reach:0-0','boost',y,x+1,interval=12,amount=4,
              rotationGroup='frontier-boost',activeSeconds=60,label={'north-cache':'North Crossing','river-cache':'Riverbank','garden-cache':'Seed Garden'}[name]),
    ])
for name,region_bounds in [('west-roam',{'x':4,'y':18,'width':21,'height':25}),
                           ('east-roam',{'x':37,'y':17,'width':9,'height':27}),
                           ('north-roam',{'x':16,'y':5,'width':12,'height':12})]:
    spawns.append(spawn(name,'willow-reach:0-0','npc',0,0,interval=18,npc='wild-bloop',
                        randomPosition=True,region=region_bounds,minPlayerDistance=5,
                        count=2,maxAlive=3,lifetimeSeconds=120))
manifest={
    'name':'Bloopworld · Willow Frontier',
    'description':'Two refuges. A living woodland frontier. A quiet place to solve things together.',
    'entry':{'stage':'willow-refuge:0-0','y':7,'x':6},
    'teams':[{'id':team,'label':label,'color':team,
              'spawn':{'stage':f'willow-refuge:0-{i}','y':7,'x':6}}
             for i,(team,label) in enumerate([('sky-blue','Sky Blue'),('fuchsia','Fuchsia')])],
    'defaultTeam':'sky-blue','maxPlayers':100,'lifecycle':'persistent',
    'npcs':[{'id':'wild-bloop','name':'Wild bloop','team':'wild',
             'appearance':'orange r1 red-b thin','health':50,'money':0,
             'intervalMs':600,'lifetimeSeconds':120,
             'program':[{'action':'chase','ticks':3,'condition':'player-nearby','radius':6},
                        {'action':'attack','ticks':1,'condition':'player-nearby','radius':1,'damage':25},
                        {'action':'wander','ticks':1,'condition':'always'}]}],
    'spawns':spawns,'achievements':[],
    'puzzles':[
        {'id':'quiet-solo','stage':'willow-solo:0-0','plates':[{'y':9,'x':5}],
         'goal':{'y':4,'x':8},'gates':[{'y':6,'x':x} for x in [7,8,9]],'minPlayers':1},
        {'id':'quiet-duo','stage':'willow-duo:0-0','plates':[{'y':9,'x':5},{'y':9,'x':11}],
         'goal':{'y':9,'x':11},'gates':[{'y':6,'x':x} for x in [7,8,9]],'minPlayers':2,'playersOnly':True},
    ],
}

# Keep this slice independent of legacy content; restore the old review tile.
legacy_space=json.loads((COLLECTION/'spaces/infirmary.json').read_text())
legacy=next(a for a in legacy_space['Areas'] if a['name']=='infirmary:0-0')
legacy['blueprint']['tiles'][9][10]={'prototypeId':'7','transformation':{}}
legacy['transports']=[t for t in legacy['transports'] if not t['destStage'].startswith('willow-')]
write('spaces/infirmary.json',legacy_space)

# Validate authored connectivity, transport round trips, terrain, and spawn sites.
by_id={p['id']:p for p in prototypes}
all_areas=[*region['Areas'],*interiors,*refuges,foyer_area,solo_area,duo_area]
by_stage={a['name']:a for a in all_areas}
def reachable(area,start,open_gates=False,blocked=frozenset()):
    grid=area['blueprint']['tiles'];seen={start};pending=deque([start])
    assert by_id[grid[start[0]][start[1]]['prototypeId']]['walkable']
    while pending:
        y,x=pending.popleft()
        for yy,xx in [(y-1,x),(y+1,x),(y,x-1),(y,x+1)]:
            if not (0<=yy<len(grid) and 0<=xx<len(grid[yy])) or (yy,xx) in seen or (yy,xx) in blocked: continue
            cell=grid[yy][xx]['prototypeId']
            if by_id[cell]['walkable'] or (open_gates and cell=='willow-annex-gate'):
                seen.add((yy,xx));pending.append((yy,xx))
    return seen
seen=reachable(region['Areas'][0],SPAWN)
for point in [*COTTAGES,(44,17),(25,38),(7,39),(35,11),(37,11),(10,28),(10,37)]:
    assert point in seen, f'Unreachable frontier landmark: {point}'
for bridge_cells in [frozenset((y,x) for y in range(23,28) for x in range(26,37)),
                     frozenset((y,x) for y in range(8,13) for x in range(28,38))]:
    assert (25,38) in reachable(region['Areas'][0],SPAWN,blocked=bridge_cells), 'Either bridge must independently connect the banks'
assert not region['Areas'][0]['safe'] and all(a['safe'] for a in all_areas[1:])
for source in all_areas:
    for exit in source['transports']:
        assert exit['destStage'] in by_stage, f'Outside-world transport: {exit}'
        destination=by_stage[exit['destStage']]
        dest_point=(exit['destY'],exit['destX'])
        assert by_id[destination['blueprint']['tiles'][exit['destY']][exit['destX']]['prototypeId']]['walkable']
        assert not any((t['sourceY'],t['sourceX'])==dest_point for t in destination['transports']), f'Immediate teleport loop: {exit}'
for i,refuge in enumerate(refuges):
    assert (3,6) in reachable(refuge,(7,6))
for puzzle in manifest['puzzles']:
    chamber=by_stage[puzzle['stage']]
    reachable_closed=reachable(chamber,(12,8))
    assert (13,8) in reachable_closed, 'Puzzle exit must never require completion'
    for plate in puzzle['plates']: assert (plate['y'],plate['x']) in reachable_closed
    goal=(puzzle['goal']['y'],puzzle['goal']['x'])
    if puzzle.get('playersOnly'):
        assert goal in [(p['y'],p['x']) for p in puzzle['plates']], 'Two players must be enough to finish the duo relay'
    else:
        assert goal not in reachable_closed and goal in reachable(chamber,(12,8),True)
for rule in spawns:
    if rule['randomPosition']:
        bounds=rule['region']
        eligible={(y,x) for y in range(bounds['y'],bounds['y']+bounds['height'])
                  for x in range(bounds['x'],bounds['x']+bounds['width'])
                  if by_id[tiles[y][x]['prototypeId']]['walkable']}
        assert len(eligible)>rule['maxAlive'] and eligible.issubset(seen), f'Inaccessible spawn region: {rule}'
        continue
    a=by_stage[rule['stage']]
    assert by_id[a['blueprint']['tiles'][rule['y']][rule['x']]['prototypeId']]['walkable'], f'Blocked spawn: {rule}'
for y,row in enumerate(tiles):
    for x,cell in enumerate(row):
        if in_snow(y,x) and by_id[cell['prototypeId']]['layer2css'].startswith('scenery-'):
            assert by_id[cell['prototypeId']]['cssColor']=='willow-snow' or 'willow-snow' in by_id[cell['prototypeId']]['layer1css'], f'Green scenery base in snow at {(y,x)}'

write('prototypes/willow-reach.json',prototypes)
write('fragments/willow-reach.json',fragments)
write('spaces/willow-reach.json',region)
write('spaces/willow-cottage.json',space('willow-cottage',interiors,16,'disconnected'))
write('spaces/willow-refuge.json',space('willow-refuge',refuges,12,'disconnected'))
write('spaces/willow-annex.json',space('willow-annex',[foyer_area],16))
write('spaces/willow-solo.json',space('willow-solo',[solo_area],16))
write('spaces/willow-duo.json',space('willow-duo',[duo_area],16))
write('interactables/willow-reach.json',[
    {'id':'willow-seed','name':'Garden seed','setName':SET,
     'cssClass':'gold r1 willow-wood-b thin','pushable':True,
     'walkable':False,'fragile':False,'reactions':''},
    {'id':'willow-puzzle-block','name':'Quiet annex block','setName':SET,
     'cssClass':'annex-edge annex-cyan-b thin r0','pushable':True,
     'walkable':False,'fragile':False,'reactions':''},
])
write('manifest.json',manifest)
palette_path=ROOT/'tools/main/data/colors/colors.json'
palette=json.loads(palette_path.read_text())
for name,color in COLORS.items():
    if not any(c['cssClassName']==name for c in palette):
        palette.append({'cssClassName':name,'R':int(color[1:3],16),'G':int(color[3:5],16),'B':int(color[5:7],16),'A':''})
palette_path.write_text(json.dumps(palette,indent=2)+'\n')
print(f'Willow Frontier: {N}×{N}, {len(trees)} trees, {len(seen)} reachable outdoor tiles, {len(prototypes)} prototypes.')
print('Two refuges; two river crossings; three cottages; solo + duo annex; 21 recurring spawn rules.')
