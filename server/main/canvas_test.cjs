const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');

test('world palettes drive static and animated canvas colors and reset between worlds', () => {
    const context = vm.createContext({
        window: { addEventListener() {} },
        document: { getElementById() { return null; } },
    });
    vm.runInContext(fs.readFileSync(`${__dirname}/assets/canvas.js`, 'utf8'), context);
    vm.runInContext(`applyCanvasPalette({ cssRules: [
        { selectorText: '.custom', style: { backgroundColor: 'rgba(80, 120, 160, 0.5)' } },
        { selectorText: '.green', style: { backgroundColor: 'rgb(12, 34, 56)' } },
        { selectorText: '.green-b', style: { borderColor: 'rgb(0, 0, 0)' } }
    ] });`, context);
    assert.equal(vm.runInContext('COLOR_MAP.custom', context), 'rgba(80, 120, 160, 0.5)');
    assert.equal(vm.runInContext('COLOR_MAP.green', context), 'rgb(12, 34, 56)');
    assert.equal(vm.runInContext('mixColor(COLOR_MAP.custom, COLOR_MAP.green, 0.5)', context), 'rgba(46, 77, 108, 0.75)');
    vm.runInContext('applyCanvasPalette({ cssRules: [] })', context);
    assert.equal(vm.runInContext('COLOR_MAP.custom', context), undefined);
    assert.equal(vm.runInContext('COLOR_MAP.green', context), 'rgb(32, 255, 60)');
});

test('scenery uses anchored depth ordering and keeps offscreen canopy anchors', () => {
    const context = vm.createContext({window:{addEventListener(){}},document:{getElementById(){return null;}},
        topLeftY:10,topLeftX:10,height:12,width:24,stageHeight:48,stageWidth:48});
    vm.runInContext(fs.readFileSync(`${__dirname}/assets/scenery.js`,'utf8'),context);
    vm.runInContext(fs.readFileSync(`${__dirname}/assets/canvas.js`,'utf8'),context);
    vm.runInContext(`stage.Lf2[24][15]='scenery-pine'; stage.Lf2[14][15]='scenery-oak'; stage.Lp1[13][15]='sky-blue'; stage.Lp1[15][15]='fuchsia';`,context);
    const items=JSON.parse(vm.runInContext(`JSON.stringify([...worldDrawables().objects,...worldDrawables().players].sort((a,b)=>a.order-b.order))`,context));
    assert.deepEqual(items.map(x=>x.kind||x.classes),['sky-blue','oak','fuchsia','pine']);
    assert.equal(vm.runInContext(`BloopScenery.kind('scenery-unknown')`,context),null);
    vm.runInContext('resetStageBuffers()',context);
    assert.equal(vm.runInContext('worldDrawables().objects.length + worldDrawables().players.length',context),0);
});

test('responsive camera dimensions stay bounded and use short landscape screens', () => {
    const context=vm.createContext({window:{addEventListener(){}},document:{addEventListener(){}}});
    vm.runInContext(fs.readFileSync(`${__dirname}/assets/game-shell.js`,'utf8'),context);
    const viewport=(w,h)=>JSON.parse(vm.runInContext(`JSON.stringify(desiredViewport(${w},${h}))`,context));
    assert.deepEqual(viewport(667,300),{rows:11,columns:24});
    assert.deepEqual(viewport(375,667),{rows:22,columns:12});
    assert.deepEqual(viewport(8000,8000),{rows:32,columns:32});
    assert.deepEqual(viewport(10,10),{rows:8,columns:8});
});

test('fullscreen failures leave a usable fallback and an accurate button state', async () => {
    const button={setAttribute(name,value){this[name]=value;}},notice={textContent:'',hidden:true};
    const context=vm.createContext({window:{addEventListener(){}},document:{documentElement:{requestFullscreen:async()=>{throw Error('denied');}},fullscreenElement:null,addEventListener(){},querySelector(){return null;},getElementById(id){return id==='fullscreen-toggle'?button:notice;}}});
    vm.runInContext(fs.readFileSync(`${__dirname}/assets/game-shell.js`,'utf8'),context);
    await vm.runInContext('toggleGameFullscreen()',context);
    assert.match(notice.textContent,/could not open/);
    assert.equal(notice.hidden,false);
    vm.runInContext('updateFullscreenControl()',context);
    assert.equal(button['aria-pressed'],'false');
    assert.equal(button['aria-label'],'Enter fullscreen');
});

test('only the viewer triggers fading, which reveals everyone beneath the same scenery', () => {
    const draws=[];
    const makeContext=name=>new Proxy({name,canvas:{width:640,height:480}}, {
        get(target,key) {return key in target ? target[key] : ()=>{};},
    });
    const main=makeContext('main');
    const canvas={getContext:()=>main}; main.canvas=canvas;
    const context=vm.createContext({
        window:{addEventListener(){}},performance:{now:()=>0},topLeftY:10,topLeftX:10,height:12,width:24,stageHeight:48,stageWidth:48,
        document:{getElementById:id=>id==='Lp1'?canvas:null},
        BloopScenery:{catalog:{oak:{width:2.8,height:3.2}},kind:classes=>classes==='scenery-oak'?'oak':null,
            draw:(ctx,kind,x,y,size,options)=>draws.push({context:ctx.name,x,alpha:options?.alpha??1,operation:ctx.globalCompositeOperation})},
    });
    vm.runInContext(fs.readFileSync(`${__dirname}/assets/canvas.js`,'utf8'),context);
    vm.runInContext(`stage.Lf2[14][15]='scenery-oak';stage.Lf2[14][20]='scenery-oak';
        stage.Lp1[13][15]='sky-blue';stage.Lp1[13][16]='sky-blue';stage.Lp1[13][20]='orange';drawWorldObjects()`,context);
    assert.deepEqual(draws.filter(x=>x.context==='main').map(x=>x.alpha),[1,1]);
    draws.length=0;
    vm.runInContext('localPlayerPosition={y:13,x:15};drawWorldObjects()',context);
    assert.deepEqual(draws.filter(x=>x.context==='main').map(x=>x.alpha),[.48,1]);
    assert.equal(draws.length,2, 'only scenery is drawn; no per-character occlusion masks');
    draws.length=0;
    vm.runInContext('localPlayerPosition={y:15,x:15};drawWorldObjects()',context);
    assert.deepEqual(draws.filter(x=>x.context==='main').map(x=>x.alpha),[1,1]);
    vm.runInContext('resetStageBuffers()',context);
    assert.equal(vm.runInContext('localPlayerPosition',context),null);
});
