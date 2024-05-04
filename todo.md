# Todo List

## Engine
- [ ] Update player view refactor
- [ ] Admin screen
- [ ] Test client

## Design Workspace
- [ ] Rotations
  - [-] New collection
  - [ ] Convert materials to protos
    - [-] Create default protos
    - [ ] Unique id 
    - [ ] Edit Proto
    - [-] define rotate(proto)
  - [-] Update fragment schema to have transformations 
    - [-] New Fragment Set
    - [-] New Fragment
    - [-] Fragment has protos
    - [-] Fragment applies transformations
  - [ ] Modify Transformations
    - [ ] Fragment Transform Proto
    - [ ] Area transform proto? 
  - [ ] Blueprint page 
    - [ ] Place fragment on blueprint
    - [ ] Transform fragment 
    - [ ] Compile blueprint (blueprint layers/actions -> defaultFragment)
  - [ ] Compile Collection 
    - [ ] DefaultFragements([][]proto) -> areas + materials


### Transformation syntax:
 "ceiling1css": "#{bg2} .{r2}@{tr} _{thick#{fg0}.{r2}@{tr}} ={hoz|vert#{black}*{2})"
 #color 
 .radius
 @orientation
 _border(#.@)
 =lines(#*quantity) @

 
