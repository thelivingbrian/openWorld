let gridCtx = {};

let cellSize = 30;

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

function getStyleForClasses(classes) {
    const tokens = classes.split(/\s+/);
    let baseColor = null;
    let alpha = 1.0;

    for (const token of tokens) {

        // 1. Transparency token? (trsp20 → alpha .80)
        const match = token.match(/^trsp(\d{2})$/);
        if (match) {
            const pct = Number(match[1]);   // 20, 40, 60, 80
            alpha = 1 - (pct / 100);
            continue;
        }

        // 2. Strip -b, -t suffixes
        const cleanToken = token.replace(/-[bt]$/, "");

        // 3. Does it map to a known color?
        if (COLOR_MAP[cleanToken]) {
            baseColor = COLOR_MAP[cleanToken];
        }
    }

    // Default: full transparency
    if (!baseColor) return "rgba(0,0,0,0)";

    // If the color is RGB, convert to RGBA with computed alpha
    if (baseColor.startsWith("rgb(")) {
        const rgb = baseColor.slice(4, -1); // strip "rgb(" and ")"
        return `rgba(${rgb}, ${alpha})`;
    }

    // If the color is already rgba, just force new alpha
    if (baseColor.startsWith("rgba(")) {
        const parts = baseColor.slice(5, -1).split(",");
        const rgb = parts.slice(0, 3).join(",");
        return `rgba(${rgb}, ${alpha})`;
    }

    // If it's a hex code, rely on canvas to apply alpha globally
    return baseColor;
}

// Draw a single cell at local (y, x) in the current viewport
function drawGridCell(id, y, x, classes) {
    if (!gridCtx[id]) {
        const gridCanvas = document.getElementById(id);
        if (!gridCanvas) {
            console.warn(`drawGridCell: no canvas with id ${id}`);
            return;
        }
        gridCtx[id] = gridCanvas.getContext("2d");
    }

    const fill = getStyleForClasses(classes);

    const ctx = gridCtx[id];
    ctx.fillStyle = fill;

    ctx.fillRect(
        x * cellSize,
        y * cellSize,
        cellSize,
        cellSize
    );
}

// Optional: full redraw if setGrid/shiftGrid change the viewport
function redrawGridFromModel() {
	// You already have the world model on the client or can request it.
	// Iterate through visible cells and call drawGridCell(y, x, classStr)
	// based on your client-side state.
}
