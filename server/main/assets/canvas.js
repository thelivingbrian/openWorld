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

let gridCtx = {};          // id -> 2d context
let cellSize   = 30;          // will match your grid / canvas size

// full world state: world[layerId][worldY][worldX] = classes string
// you size this however big your world is
const maxStageHeight = 256;
const maxStageWidth  = 256;

let stage = {};
for (const id of layerIds) {
    stage[id] = Array.from({ length: maxStageHeight }, () =>
        Array.from({ length: maxStageWidth }, () => "")
    );
}

function getCtx(id) {
    if (!gridCtx[id]) {
        const canvas = document.getElementById(id);
        gridCtx[id]  = canvas.getContext("2d");
    }
    return gridCtx[id];
}

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
    "black": "rgb(0, 0, 0)",
    "med-gray": "rgb(145, 145, 145)",
    "dark-lavender": "rgb(172, 152, 219)",
};

function getColorAndAlphaForClasses(classes) {
    const tokens = classes.split(/\s+/);

    let baseColor = null;
    let alpha = 1.0;

    for (const token of tokens) {
        // Handle transparency marker: trsp20, trsp40, trsp60, trsp80
        const trsp = token.match(/^trsp(\d{2})$/);
        if (trsp) {
            const pct = Number(trsp[1]);     // e.g. 20
            alpha = pct / 100.0;            // visibility: 0.2, 0.4, 0.6, 0.8
            continue;
        }

        // Strip -b / -t variants; we only care about base color class
        const clean = token.replace(/-[bt]$/, "");

        if (COLOR_MAP[clean]) {
            baseColor = COLOR_MAP[clean];
        }
    }

    // Special case: invisible class
    if (tokens.includes("invisible")) {
        return { baseColor: null, alpha: 0 }; // draw function will clearRect
    }

    // No recognized color → treat as fully transparent (don't draw)
    if (!baseColor) {
        return { baseColor: null, alpha: 0 };
    }

    return { baseColor, alpha };
}

function drawGridCell(id, y, x, classes) {
    const ctx = getCtx(id);
    if (!ctx) return;

    const { baseColor, alpha } = getColorAndAlphaForClasses(classes);

    const px = x * cellSize;
    const py = y * cellSize;

    // If no color or explicitly invisible → clear this tile on this canvas
    if (!baseColor || classes.includes("invisible")) {
        ctx.clearRect(px, py, cellSize, cellSize);
        return;
    }

    // Erase old contents of *this tile* before drawing new color
    ctx.clearRect(px, py, cellSize, cellSize);

    ctx.save();
    ctx.globalAlpha = alpha;
    ctx.fillStyle = baseColor;
    ctx.fillRect(px, py, cellSize, cellSize);
    ctx.restore();
}

function redrawStage() {
    for (const id of layerIds) {
        const ctx    = getCtx(id);
        const canvas = ctx.canvas;

        // clear this whole layer
        ctx.clearRect(0, 0, canvas.width, canvas.height);

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
