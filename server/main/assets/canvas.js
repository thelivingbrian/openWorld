const layerIds = [
    "Lg1", "Lg2",
    "Lf1", "Lf2",
    "Lp1",
    "Li1",
    "Ls1",
    "Lc1", "Lc2",
    "Lw1",
    "Lt1",
];

const COLOR_MAP = {
    "invisible": "rgba(0, 0, 0, 0)",
    "day": "rgb(170, 255, 255)",
    "night": "rgb(18, 5, 0)",
    "twilight": "rgb(255, 106, 92)",
    "green": "rgb(32, 255, 60)",
    "dark-green": "rgb(32, 155, 60)",
    "lime": "rgb(150, 255, 60)",
    "yellow": "rgb(255, 255, 60)",
    "orange": "rgb(255, 150, 60)",
    "red": "rgb(255, 32, 60)",
    "dark-red": "rgb(139, 0, 0)",
    "blue": "rgb(72, 52, 238)",
    "fuchsia": "rgb(253, 52, 172)",
    "dim-fuchsia": "rgb(89, 18, 60)",
    "pink": "rgb(253, 182, 215)",
    "salmon": "rgb(255, 165, 145)",
    "half-gray": "rgb(120, 120, 129)",
    "dark-blue": "rgb(32, 15, 65)",
    "sand": "rgb(250, 245, 175)",
    "grass": "rgb(210, 250, 180)",
    "ice": "rgb(225, 240, 255)",
    "wall": "rgb(100, 100, 100)",
    "white": "rgb(255, 255, 255)",
    "light-gray": "rgb(210, 210, 210)",
    "cinnamon": "rgb(210, 125, 45)",
    "nude": "rgb(242, 210, 189)",
    "chocolate": "rgb(123, 63, 0)",
    "tan": "rgb(210, 180, 140)",
    "copper": "rgb(184, 115, 51)",
    "dark-gray": "rgb(60, 60, 60)",
    "purple": "rgb(128, 0, 128)",
    "teal": "rgb(0, 128, 128)",
    "turquoise": "rgb(64, 224, 208)",
    "lavender": "rgb(218, 203, 250)",
    "olive": "rgb(128, 128, 0)",
    "gold": "rgb(255, 215, 0)",
    "navy": "rgb(0, 0, 128)",
    "peach": "rgb(255, 218, 185)",
    "burgundy": "rgb(128, 0, 32)",
    "dark-grass": "rgb(160, 200, 130)",
    "sky-blue": "rgb(48, 152, 240)",
    "dim-sky-blue": "rgb(17, 53, 84)",
    "black": "rgb(0, 0, 0)",
    "med-gray": "rgb(145, 145, 145)",
    "dark-lavender": "rgb(172, 152, 219)",
};

const BORDER_WIDTH_MAP = {
    thin: 1,
    med:  2,
    thick: 5,
};

let canvasLayers = {};          
let cellSize   = 30;          

const maxStageHeight = 256;
const maxStageWidth  = 256;

let stage = {};
for (const id of layerIds) {
    stage[id] = Array.from({ length: maxStageHeight }, () =>
        Array.from({ length: maxStageWidth }, () => "")
    );
}

function getCanvasContextByLayerId(id) {
    if (!canvasLayers[id]) {
        const canvas = document.getElementById(id);
        canvasLayers[id]  = canvas.getContext("2d");
    }
    return canvasLayers[id];
}

// Redraw entire stage

function redrawStage() {
    for (const id of layerIds) {
        const ctx    = getCanvasContextByLayerId(id);
        const canvas = ctx.canvas;

        // clear this whole layer
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        if (!stage[id]) {
            stage[id] = Array.from({ length: maxStageHeight }, () =>
                Array.from({ length: maxStageWidth }, () => "")
            );
            continue;
        }
        // draw visible tiles for this layer
        for (let vy = 0; vy < height; vy++) {
            for (let vx = 0; vx < width; vx++) {
                const wy = topLeftY + vy;
                const wx = topLeftX + vx;

                // bounds check world
                if (wy < 0 || wy >= maxStageHeight || wx < 0 || wx >= maxStageWidth) {
                    continue;
                }

                const classes = stage[id][wy][wx];
                if (!classes) continue;

                drawGridCell(id, vy, vx, classes);
            }
        }
    }
}

// Draw a single cell

function drawGridCell(id, y, x, classes) {
    const ctx = getCanvasContextByLayerId(id);
    if (!ctx) return;

    const { fillColor, strokeColor, alpha, borderWidth } =
        getDrawingStyle(classes);

    const px = x * cellSize;
    const py = y * cellSize;

    // Clear the tile first
    ctx.clearRect(px, py, cellSize, cellSize);

    if (!fillColor && !strokeColor) return;

    const radii = getCornerRadii(classes, cellSize);

    const inset   = borderWidth;
    const innerX  = px + inset;
    const innerY  = py + inset;
    const innerW  = cellSize - inset * 2;
    const innerH  = cellSize - inset * 2;
    const innerR  = radii.map(r => Math.max(0, r - inset));

    ctx.save();
    ctx.globalAlpha = alpha;  // trspXX applies to both border and fill (like CSS opacity)

    // 1) Fill interior (background) if there is a fillColor
    if (fillColor && innerW > 0 && innerH > 0) {
        ctx.fillStyle = fillColor;
        ctx.beginPath();
        pathRoundedRect(ctx, innerX, innerY, innerW, innerH, innerR);
        ctx.fill();
    }

    // 2) Draw border as a ring (outer − inner, using even-odd rule)
    if (strokeColor && borderWidth > 0) {
        ctx.fillStyle = strokeColor;
        ctx.beginPath();
        // outer path
        pathRoundedRect(ctx, px, py, cellSize, cellSize, radii);

        if (innerW > 0 && innerH > 0) {
            // inner path (hole)
            pathRoundedRect(ctx, innerX, innerY, innerW, innerH, innerR);
            ctx.fill("evenodd");   // fill ring only
        } else {
            // very small tiles / huge border: just fill outer shape
            ctx.fill();
        }
    }

    ctx.restore();
}



function getCornerRadii(classes, cellSize) {
    const tokens = classes.split(/\s+/);

    // TL, TR, BR, BL
    const radii = [0, 0, 0, 0];

    // base radius percent if r0/r1/r2 present
    let basePct = 0;

    // first pass: find base r0 / r1 / r2
    for (const token of tokens) {
        if (token === "r0") basePct = 0.25;
        if (token === "r1") basePct = 0.50;
        if (token === "r2") basePct = 0.75;
    }

    if (basePct > 0) {
        const baseRadius = basePct * cellSize;
        radii[0] = baseRadius; // TL
        radii[1] = baseRadius; // TR
        radii[2] = baseRadius; // BR
        radii[3] = baseRadius; // BL
    }

    // second pass: per-corner overrides like r0-bl, r1-tr, etc.
    for (const token of tokens) {
        const m = token.match(/^r([0-2])-(tl|tr|br|bl)$/);
        if (!m) continue;

        const level = Number(m[1]); // 0,1,2 → 25,50,75
        const corner = m[2];

        const pct = (level === 0 ? 0.25 : level === 1 ? 0.5 : 0.75);
        const r = pct * cellSize;

        const idx =
            corner === "tl" ? 0 :
            corner === "tr" ? 1 :
            corner === "br" ? 2 : 3; // bl
        radii[idx] = r;
    }

    return radii;
}

function getDrawingStyle(classes) {
    const tokens = classes.split(/\s+/);

    let fillColor   = null;
    let strokeColor = null;
    let alpha       = 1.0;
    let borderWidth = 0;

    // handle "invisible" explicitly
    if (tokens.includes("invisible")) {
        return { fillColor: null, strokeColor: null, alpha: 0, borderWidth: 0 };
    }

    for (const token of tokens) {
        // 1. transparency: trsp20, trsp40, trsp60, trsp80
        const trsp = token.match(/^trsp(\d{2})$/);
        if (trsp) {
            const pct = Number(trsp[1]); // e.g. 20 → 0.20 visible
            alpha = pct / 100.0;
            continue;
        }

        // 2. border thickness: thin, med, thick
        if (BORDER_WIDTH_MAP[token]) {
            borderWidth = BORDER_WIDTH_MAP[token];
            continue;
        }

        // 3. color classes: like "grass", "grass-b", "grass-t"
        const m = token.match(/^(.+?)(-[bt])?$/);
        if (!m) continue;

        const base   = m[1];      // "grass"
        const suffix = m[2] || ""; // "", "-b", "-t"

        const baseColor = COLOR_MAP[base];
        if (!baseColor) continue;

        if (suffix === "-b") {
            strokeColor = baseColor;
        } else if (suffix === "-t") {
            // text color; we ignore this for tiles
            continue;
        } else {
            // plain "grass" → fill color
            fillColor = baseColor;
        }
    }

    // If we have a stroke but no fill, that's okay (border-only tile)
    // If we have neither, treat as fully transparent
    if (!fillColor && !strokeColor) {
        return { fillColor: null, strokeColor: null, alpha: 0, borderWidth: 0 };
    }

    return { fillColor, strokeColor, alpha, borderWidth };
}

function pathRoundedRect(ctx, x, y, w, h, radii) {
    const [rtl, rtr, rbr, rbl] = radii.map(r => Math.max(0, r));

    // Use native roundRect if present
    if (typeof ctx.roundRect === "function") {
        ctx.roundRect(x, y, w, h, [rtl, rtr, rbr, rbl]);
        return;
    }

    // Manual path
    ctx.beginPath();
    ctx.moveTo(x + rtl, y);
    ctx.lineTo(x + w - rtr, y);
    ctx.quadraticCurveTo(x + w, y, x + w, y + rtr);

    ctx.lineTo(x + w, y + h - rbr);
    ctx.quadraticCurveTo(x + w, y + h, x + w - rbr, y + h);

    ctx.lineTo(x + rbl, y + h);
    ctx.quadraticCurveTo(x, y + h, x, y + h - rbl);

    ctx.lineTo(x, y + rtl);
    ctx.quadraticCurveTo(x, y, x + rtl, y);
    ctx.closePath();
}
