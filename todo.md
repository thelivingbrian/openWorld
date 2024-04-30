# Todo List

## Engine
- [ ] Update player view refactor
- [ ] Admin screen
- [ ] Test client

## Design Workspace
- [ ] Rotations
  - [-] New collection
  - [ ] Convert materials to protos
    - [] Unique id 
    - [] define rotate(proto)
  - [ ] Update fragment schema to have transformations 
  - [ ] Blueprint page 
    - [ ] Place fragment on blueprint
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

 
