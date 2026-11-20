/* Shared, resolution-independent scenery for gameplay and the world editor.
 * Anchors are tile-centre / tile-bottom. Extents are in square world tiles;
 * collision lives in the source fragment, never in the painted silhouette. */
(function (root) {
    const catalog = {
        oak: { width: 2.8, height: 3.2 }, pine: { width: 2.4, height: 3.5 },
        birch: { width: 2.3, height: 3.1 }, cottage: { width: 5, height: 4 },
        lantern: { width: 1, height: 1.9 }, sign: { width: 1.5, height: 1.3 },
        rock: { width: 1.4, height: 1 }, flower: { width: 1, height: .75 },
        bench: { width: 2, height: 1.1 }, well: { width: 2.6, height: 1.8 },
    };
    function kind(classes) {
        const match = /(?:^|\s)scenery-([a-z]+)(?:\s|$)/.exec(classes || '');
        return match && catalog[match[1]] ? match[1] : null;
    }
    function draw(ctx, type, x, y, size, options = {}) {
        if (!catalog[type]) return;
        ctx.save(); ctx.translate(x, y); ctx.scale(size, size);
        ctx.globalAlpha = options.alpha ?? 1;
        const ellipse = (x, y, rx, ry, color) => {
            ctx.fillStyle = color; ctx.beginPath(); ctx.ellipse(x, y, rx, ry, 0, 0, Math.PI * 2); ctx.fill();
        };
        const box = (x, y, w, h, r, color) => {
            ctx.fillStyle = color; ctx.beginPath(); ctx.roundRect(x, y, w, h, r); ctx.fill();
        };
        const poly = (points, color) => {
            ctx.fillStyle = color; ctx.beginPath(); points.forEach(([x,y],i) => i ? ctx.lineTo(x,y) : ctx.moveTo(x,y)); ctx.closePath(); ctx.fill();
        };
        const shadow = (w, h = .2) => ellipse(.2, -.08, w, h, '#203f392b');
        if (type === 'oak' || type === 'birch') {
            shadow(1.1, .28);
            const birch = type === 'birch';
            box(-.18, -1.55, .36, 1.5, .09, birch ? '#e7dcc0' : '#805c43');
            box(-.12, -1.5, .1, 1.4, .04, birch ? '#fff0cf' : '#b18a5e');
            if (birch) for(let i=0;i<3;i++) box(-.16,-.35-i*.33,.2,.055,.02,'#7c7966');
            poly([[-.06,-.65],[-.55,-1.45],[-.4,-1.5],[.1,-.9]], birch ? '#c4c2a2' : '#805c43');
            const dark = birch ? '#578a64' : '#326c57';
            const mid = birch ? '#83ae76' : '#4c9269';
            const light = birch ? '#b6ce88' : '#7fb074';
            ellipse(0,-1.9,birch?.95:1.28,.94,dark);
            ellipse(-.65,-1.85,.65,.65,dark); ellipse(.6,-1.8,.7,.64,dark);
            ellipse(-.25,-2.42,.83,.76,mid); ellipse(.48,-2.25,.68,.65,mid);
            ellipse(-.6,-2.12,.63,.64,mid); ellipse(.13,-1.96,.7,.64,mid);
            ellipse(-.33,-2.65,.55,.39,light); ellipse(-.77,-2.2,.28,.22,light);
            ellipse(.53,-2.4,.18,.13,'#b2ca8b');
            box(-.55,-1.64,.35,.08,.04,'#25544538');
        } else if (type === 'pine') {
            shadow(.95,.24); box(-.15,-1.1,.3,1.05,.06,'#8d7556');
            for (let i=0;i<3;i++) {
                const yy=-.65-i*.8, w=1.13-i*.25;
                ctx.fillStyle=i===0?'#286653':i===1?'#367c60':'#559575';
                ctx.beginPath(); ctx.moveTo(0,yy-1.25); ctx.quadraticCurveTo(w*.5,yy-.55,w,yy-.12);
                ctx.quadraticCurveTo(w*1.1,yy+.2,0,yy+.22); ctx.quadraticCurveTo(-w*1.1,yy+.2,-w,yy-.12);
                ctx.quadraticCurveTo(-w*.45,yy-.55,0,yy-1.25); ctx.fill();
                poly([[0,yy-1.22],[-w*.75,yy-.1],[-w*.25,yy-.28]],'#a8cba038');
            }
        } else if (type === 'cottage') {
            shadow(2.5,.35);
            box(-2.12,-1.96,4.24,1.9,.1,'#e4d8ac');
            box(1.58,-1.96,.54,1.9,.05,'#c5b98f');
            box(-2.15,-.25,4.3,.2,.04,'#a6a58a');
            box(-1.8,-3.6,.55,1.1,.06,'#bd8267'); box(-1.89,-3.64,.72,.18,.04,'#e7b396');
            poly([[-2.5,-1.8],[-1.55,-3.55],[1.43,-3.55],[2.5,-1.8]],'#914e4b');
            poly([[-2.47,-1.9],[-1.54,-3.55],[1.35,-3.55],[2.22,-1.9]],'#cb7967');
            poly([[-2.42,-1.92],[-1.5,-3.5],[-1.26,-3.5],[-2.13,-1.92]],'#e4a080');
            for(let i=0;i<3;i++) box(-1.8-i*.19,-3.12+i*.43,3.58+i*.38,.045,.02,'#a45e534d');
            box(-2.47,-1.94,4.95,.16,.07,'#854c48');
            box(-.49,-1.45,.98,1.45,.12,'#967351'); box(-.34,-1.32,.68,1.27,.08,'#557c6d');
            box(-.22,-1.18,.44,.42,.06,'#a7d6cb'); ellipse(.21,-.57,.055,.055,'#f4ce79');
            box(-.62,-.08,1.24,.18,.05,'#c1bba0');
            for (const xx of [-1.38,1.31]) {
                box(xx-.34,-1.45,.68,.73,.12,'#f7e9c4'); box(xx-.25,-1.36,.5,.53,.07,'#7eaaa4');
                box(xx-.025,-1.38,.05,.57,0,'#e5dcc0'); box(xx-.27,-1.12,.54,.05,0,'#e5dcc0');
                box(xx-.42,-.72,.84,.14,.03,'#ae7f5d');
                for(let j=0;j<4;j++) ellipse(xx-.27+j*.18,-.75,.1,.13,j%2?'#eeb99b':'#5c8c61');
            }
        } else if (type === 'well') {
            shadow(1.3,.36);
            box(-1.15,-.9,2.3,.75,.15,'#8f9d88');
            ellipse(0,-.3,1.15,.35,'#8f9d88');
            for(let i=0;i<6;i++) box(-.95+i*.37,-.65,.025,.4,.01,'#6c81713a');
            ellipse(0,-.96,1.2,.62,'#d0cfb1');
            ellipse(0,-1, .91,.4,'#4e939c');
            ellipse(-.1,-1.07,.67,.22,'#83b6b2');
            box(-.16,-1.57,.32,.62,.08,'#b4bda3');
            ellipse(0,-1.57,.27,.14,'#e3dcc0');
            ellipse(0,-1.62,.14,.07,'#a6d0c4');
        } else if (type === 'lantern') {
            shadow(.4,.12); box(-.06,-1.55,.12,1.5,.04,'#5d6d5c');
            box(-.25,-1.65,.5,.55,.09,'#65735d'); box(-.16,-1.55,.32,.34,.06,'#ffe4a1');
            poly([[-.33,-1.65],[0,-1.88],[.33,-1.65]],'#4d6559');
            ellipse(0,-1.38,.38,.4,'#ffe4a118');
        } else if (type === 'sign') {
            shadow(.5,.15); box(-.07,-.8,.14,.78,.04,'#8c7956');
            box(-.65,-1.22,1.3,.58,.1,'#846b4e'); box(-.59,-1.17,1.18,.43,.07,'#d5bd87');
            box(-.4,-1.03,.7,.055,.025,'#7d7f5f'); box(-.4,-.9,.45,.045,.02,'#7d7f5f');
        } else if (type === 'rock') {
            shadow(.64,.18); ellipse(0,-.34,.65,.38,'#7f9485'); ellipse(-.12,-.47,.49,.35,'#aab9a1');
            ellipse(-.2,-.63,.25,.12,'#c5d0b5');
        } else if (type === 'bench') {
            shadow(.9,.19); box(-.7,-.5,.12,.47,.03,'#667a63'); box(.58,-.5,.12,.47,.03,'#667a63');
            box(-.87,-1,1.74,.32,.06,'#be9262'); box(-.9,-.55,1.8,.23,.05,'#d0ac76');
            box(-.9,-.34,1.8,.08,.02,'#9c7c55');
        } else if (type === 'flower') {
            for (let i=0;i<3;i++) {
                const xx=(i-1)*.25, yy=-.25-(i%2)*.2;
                box(xx-.025,yy,.05,-yy,.02,'#5e8e65');
                ellipse(xx,yy,.12,.1,i%2?'#f0c985':'#f2dac1'); ellipse(xx,yy,.04,.04,'#cc9c59');
            }
        }
        ctx.restore();
    }
    root.BloopScenery = { catalog, kind, draw };
    if (typeof module !== 'undefined') module.exports = root.BloopScenery;
})(typeof globalThis !== 'undefined' ? globalThis : window);
