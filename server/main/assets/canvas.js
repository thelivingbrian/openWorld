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

const WEATHER_MODE_RENDERERS = {
    raining: drawRainingWeatherFrame,
};

const weatherRuntime = {
    drops: [],
    lastFrameAtMs: 0,
    seededMode: "",
    seededWidth: 0,
    seededHeight: 0,
};

let visualEffectsAnimationStarted = false;

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

function clearCanvasAndStage() {
    canvasLayers = {};
    stage = {};
} 

function getCanvasContextByLayerId(id) {
    if (!canvasLayers[id]) {
        const canvas = document.getElementById(id);
        if (!canvas) return null;
        canvasLayers[id]  = canvas.getContext("2d");
    }
    return canvasLayers[id];
}

function resizeCanvas() {
    const game = document.getElementById("game");
    if (!game) return;

    const dpr = 1; // don't even mess with window.devicePixelRatio -> is land mine 
    const vw  = window.innerWidth;
    const vh  = window.innerHeight;

    // mimics previous CSS: #screen is 40% wide in landscape, 87% in portrait
    const landscape = vw >= vh;
    const maxScreenWidth = landscape ? vw * 0.40 : vw * 0.87;
    const maxScreenHeight = vh; // we don't restrict height as much

    const avail = Math.min(maxScreenWidth, maxScreenHeight);

    // tiles across (assuming square viewport); use width or height as needed
    const maxCellsAcross = width;

    // snap cell size to an integer so tiles land on whole pixels
    const newCellSize = Math.max(1, Math.floor(avail / maxCellsAcross));
    const stageCssSize = newCellSize * maxCellsAcross;  // exact integer

    // update globals
    cellSize = newCellSize;

    // size #game container
    game.style.width  = `${stageCssSize}px`;
    game.style.height = `${stageCssSize}px`;

    // Assuming square canvases, Todo: DPR adds no value here 
    const backingWidth = Math.round(stageCssSize * dpr);
    const backingHeight = Math.round(stageCssSize * dpr);

    // resize each canvas backing store + CSS size
    for (const id of layerIds) {
        const canvas = document.getElementById(id);
        if (!canvas) continue;

        canvas.width  = backingWidth;
        canvas.height = backingHeight;

        canvas.style.width  = `${stageCssSize}px`;
        canvas.style.height = `${stageCssSize}px`;

        const ctx = getCanvasContextByLayerId(id);
        if (!ctx) continue;

        // normalize 0..stageCssSize coordinates in CSS pixels
        ctx.setTransform(1, 0, 0, 1, 0, 0);
    }

    redrawStage();
}

// Dynamically resize canvas
window.addEventListener("resize", resizeCanvas);

/////////////////////////////////////////////////////////////////
//  Drawing 

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

                drawGridCell(id, vy, vx, classes, undefined, wy, wx);
            }
        }
    }

    ensureVisualEffectsAnimationLoop();
}

function redrawWeatherLayerBase() {
    const id = "Lw1";
    const ctx = getCanvasContextByLayerId(id);
    if (!ctx) return;

    const canvas = ctx.canvas;
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    if (!stage[id]) {
        stage[id] = Array.from({ length: maxStageHeight }, () =>
            Array.from({ length: maxStageWidth }, () => "")
        );
        return;
    }

    for (let vy = 0; vy < height; vy++) {
        for (let vx = 0; vx < width; vx++) {
            const wy = topLeftY + vy;
            const wx = topLeftX + vx;

            if (wy < 0 || wy >= maxStageHeight || wx < 0 || wx >= maxStageWidth) {
                continue;
            }

            const classes = stage[id][wy][wx];
            if (!classes) continue;

            drawGridCell(id, vy, vx, classes, undefined, wy, wx);
        }
    }
}

function ensureVisualEffectsAnimationLoop() {
    if (visualEffectsAnimationStarted) return;

    visualEffectsAnimationStarted = true;
    requestAnimationFrame(stepVisualEffects);
}

function stepVisualEffects(nowMs) {
    const mode = getVisibleWeatherMode();
    const dynamicLayerIds = getVisibleDynamicLayerIds();
    const hasAnimatedTiles = dynamicLayerIds.length > 0;

    if (!mode && !hasAnimatedTiles) {
        weatherRuntime.lastFrameAtMs = 0;
        weatherRuntime.seededMode = "";
        weatherRuntime.drops = [];
        requestAnimationFrame(stepVisualEffects);
        return;
    }

    const dt = weatherRuntime.lastFrameAtMs === 0
        ? (1 / 60)
        : Math.min(0.05, (nowMs - weatherRuntime.lastFrameAtMs) / 1000);

    weatherRuntime.lastFrameAtMs = nowMs;

    if (hasAnimatedTiles) {
        redrawDynamicLayers(dynamicLayerIds, nowMs);
    }

    const renderMode = WEATHER_MODE_RENDERERS[mode];
    if (renderMode) {
        renderMode(dt, nowMs);
    } else {
        weatherRuntime.seededMode = "";
        weatherRuntime.drops = [];
    }

    requestAnimationFrame(stepVisualEffects);
}

function redrawDynamicLayers(layerIds, nowMs) {
    for (const id of layerIds) {
        const ctx = getCanvasContextByLayerId(id);
        if (!ctx) continue;

        for (let vy = 0; vy < height; vy++) {
            for (let vx = 0; vx < width; vx++) {
                const wy = topLeftY + vy;
                const wx = topLeftX + vx;
                if (wy < 0 || wy >= maxStageHeight || wx < 0 || wx >= maxStageWidth) {
                    continue;
                }

                const classes = stage[id][wy][wx];
                if (!classes || !hasDynamicTileToken(classes)) continue;

                drawGridCell(id, vy, vx, classes, nowMs, wy, wx);
            }
        }
    }
}

function getVisibleDynamicLayerIds() {
    const out = [];

    for (const id of layerIds) {
        if (id === "Ls1" || id === "Lw1") continue;
        if (!stage[id]) continue;

        let found = false;
        for (let vy = 0; vy < height && !found; vy++) {
            for (let vx = 0; vx < width; vx++) {
                const wy = topLeftY + vy;
                const wx = topLeftX + vx;
                if (wy < 0 || wy >= maxStageHeight || wx < 0 || wx >= maxStageWidth) {
                    continue;
                }

                const classes = stage[id][wy][wx];
                if (classes && hasDynamicTileToken(classes)) {
                    out.push(id);
                    found = true;
                    break;
                }
            }
        }
    }

    return out;
}

function getVisibleWeatherMode() {
    const weatherLayer = stage["Lw1"];
    if (!weatherLayer) return "";

    for (let vy = 0; vy < height; vy++) {
        for (let vx = 0; vx < width; vx++) {
            const wy = topLeftY + vy;
            const wx = topLeftX + vx;

            if (wy < 0 || wy >= maxStageHeight || wx < 0 || wx >= maxStageWidth) {
                continue;
            }

            const classes = weatherLayer[wy][wx];
            if (!classes) continue;

            const tokens = classes.split(/\s+/);
            for (const token of tokens) {
                if (WEATHER_MODE_RENDERERS[token]) {
                    return token;
                }
            }
        }
    }

    return "";
}

function drawRainingWeatherFrame(dtSeconds) {
    const ctx = getCanvasContextByLayerId("Lw1");
    if (!ctx) return;

    const canvas = ctx.canvas;
    const canvasWidth = canvas.width;
    const canvasHeight = canvas.height;
    const targetDropCount = Math.max(48, Math.floor((canvasWidth * canvasHeight) / 4200));

    if (
        weatherRuntime.seededMode !== "raining" ||
        weatherRuntime.seededWidth !== canvasWidth ||
        weatherRuntime.seededHeight !== canvasHeight ||
        weatherRuntime.drops.length !== targetDropCount
    ) {
        weatherRuntime.seededMode = "raining";
        weatherRuntime.seededWidth = canvasWidth;
        weatherRuntime.seededHeight = canvasHeight;
        weatherRuntime.drops = Array.from(
            { length: targetDropCount },
            () => createRainDrop(canvasWidth, canvasHeight)
        );
    }

    redrawWeatherLayerBase();

    ctx.save();
    ctx.strokeStyle = COLOR_MAP["white"];
    ctx.lineWidth = Math.max(1, Math.round(cellSize * 0.06));
    ctx.lineCap = "round";

    for (const drop of weatherRuntime.drops) {
        drop.x += drop.wind * dtSeconds;
        drop.y += drop.speed * dtSeconds;

        if (drop.y > canvasHeight + drop.length) {
            drop.y = -drop.length;
            drop.x = Math.random() * canvasWidth;
        }
        if (drop.x > canvasWidth + drop.length) {
            drop.x = -drop.length;
        }
        if (drop.x < -drop.length) {
            drop.x = canvasWidth + drop.length;
        }

        ctx.globalAlpha = drop.alpha;
        ctx.beginPath();
        ctx.moveTo(drop.x, drop.y);
        ctx.lineTo(drop.x + drop.wind * 0.08, drop.y + drop.length);
        ctx.stroke();
    }

    ctx.restore();
}

function createRainDrop(canvasWidth, canvasHeight) {
    const baseLength = Math.max(8, cellSize * 0.35);
    return {
        x: Math.random() * canvasWidth,
        y: Math.random() * canvasHeight,
        length: baseLength + Math.random() * (baseLength * 0.75),
        speed: Math.max(160, cellSize * 18) + Math.random() * (cellSize * 8),
        wind: (Math.random() * cellSize * 2.6) + (cellSize * 0.8),
        alpha: 0.2 + (Math.random() * 0.35),
    };
}

function drawGridCell(id, y, x, classes, nowMs, worldY, worldX) {
    const ctx = getCanvasContextByLayerId(id);
    if (!ctx) return;

    const renderTimeMs = nowMs ?? performance.now();
    const tileWorldY = worldY ?? (topLeftY + y);
    const tileWorldX = worldX ?? (topLeftX + x);

    const { fillColor, strokeColor, alpha, borderWidth } =
        getDrawingStyle(classes, renderTimeMs, tileWorldY, tileWorldX);

    const px = Math.round(x * cellSize);
    const py = Math.round(y * cellSize);

    // Clear the tile first
    ctx.clearRect(px, py, cellSize, cellSize);

    if (!fillColor && !strokeColor) return;

    const shape = getShapeModifiers(classes, cellSize);
    const outerX = px + shape.offsetX;
    const outerY = py + shape.offsetY;
    const outerW = shape.width;
    const outerH = shape.height;

    const radii = getCornerRadii(classes, cellSize);

    const borderTop    = shape.showBorderTop ? borderWidth : 0;
    const borderRight  = shape.showBorderRight ? borderWidth : 0;
    const borderBottom = shape.showBorderBottom ? borderWidth : 0;
    const borderLeft   = shape.showBorderLeft ? borderWidth : 0;

    const innerX  = outerX + borderLeft;
    const innerY  = outerY + borderTop;
    const innerW  = outerW - borderLeft - borderRight;
    const innerH  = outerH - borderTop - borderBottom;
    const innerR  = [
        Math.max(0, radii[0] - Math.max(borderLeft, borderTop)),
        Math.max(0, radii[1] - Math.max(borderRight, borderTop)),
        Math.max(0, radii[2] - Math.max(borderRight, borderBottom)),
        Math.max(0, radii[3] - Math.max(borderLeft, borderBottom)),
    ];

    const hasSelectiveBorder =
        borderTop !== borderWidth ||
        borderRight !== borderWidth ||
        borderBottom !== borderWidth ||
        borderLeft !== borderWidth;

    ctx.save();
    ctx.globalAlpha = alpha;  // trspXX applies to both border and fill (like CSS opacity)

    // 1) Fill interior if there is a fillColor
    if (fillColor && innerW > 0 && innerH > 0) {
        ctx.fillStyle = fillColor;
        ctx.beginPath();
        pathRoundedRect(ctx, innerX, innerY, innerW, innerH, innerR);
        ctx.fill();
    }

    // 2) Draw border as a ring (outer − inner, using even-odd rule)
    if (strokeColor && borderWidth > 0) {
        ctx.fillStyle = strokeColor;
        if (!hasSelectiveBorder) {
            // one path - multiple subpaths
            ctx.beginPath();
            // outer path
            pathRoundedRect(ctx, outerX, outerY, outerW, outerH, radii);

            if (innerW > 0 && innerH > 0) {
                // inner path (hole)
                pathRoundedRect(ctx, innerX, innerY, innerW, innerH, innerR);
                ctx.fill("evenodd");   // fill ring only
            } else {
                // very small tiles / huge border: just fill outer shape
                ctx.fill();
            }
        } else {
            drawSelectiveBorders(
                ctx,
                outerX,
                outerY,
                outerW,
                outerH,
                borderTop,
                borderRight,
                borderBottom,
                borderLeft
            );
        }
    }

    ctx.restore();

    drawSparkleOverlay(ctx, classes, outerX, outerY, outerW, outerH, alpha, renderTimeMs, tileWorldY, tileWorldX);
}

function drawSparkleOverlay(ctx, classes, x, y, w, h, alpha, timeMs, worldY, worldX) {
    if (!classes.includes("sparkle")) return;

    const t = timeMs / 1000;
    const sparkleCount = 2;

    ctx.save();
    ctx.fillStyle = COLOR_MAP["white"];

    for (let i = 0; i < sparkleCount; i++) {
        const phase = hashFloat(worldY, worldX, i, 1) * Math.PI * 2;
        const twinkle = Math.max(0, Math.sin(t * 4 + phase));
        if (twinkle < 0.35) continue;

        const px = x + w * (0.15 + (hashFloat(worldY, worldX, i, 2) * 0.7));
        const py = y + h * (0.15 + (hashFloat(worldY, worldX, i, 3) * 0.7));
        const radius = Math.max(1, cellSize * 0.05 * (0.8 + hashFloat(worldY, worldX, i, 4)));

        ctx.globalAlpha = alpha * twinkle * 0.9;
        ctx.beginPath();
        ctx.arc(px, py, radius, 0, Math.PI * 2);
        ctx.fill();
    }

    ctx.restore();
}

function getShapeModifiers(classes, cellSize) {
    const tokens = classes.split(/\s+/);

    const isHorizontalStrip = tokens.includes("s-hoz");
    const isVerticalStrip = tokens.includes("s-vert");

    let width = cellSize;
    let height = cellSize;

    if (isHorizontalStrip) {
        height = Math.max(1, Math.round(cellSize * 0.5));
    }
    if (isVerticalStrip) {
        width = Math.max(1, Math.round(cellSize * 0.5));
    }

    return {
        offsetX: Math.round((cellSize - width) / 2),
        offsetY: Math.round((cellSize - height) / 2),
        width,
        height,
        showBorderLeft: !tokens.includes("no-lr"),
        showBorderRight: !tokens.includes("no-lr"),
        showBorderTop: !tokens.includes("no-tb"),
        showBorderBottom: !tokens.includes("no-tb"),
    };
}

function drawSelectiveBorders(ctx, x, y, w, h, top, right, bottom, left) {
    if (top > 0) {
        ctx.fillRect(x, y, w, top);
    }
    if (bottom > 0) {
        ctx.fillRect(x, y + h - bottom, w, bottom);
    }
    if (left > 0) {
        ctx.fillRect(x, y, left, h);
    }
    if (right > 0) {
        ctx.fillRect(x + w - right, y, right, h);
    }
}

function getCornerRadii(classes, cellSize) {
    const tokens = classes.split(/\s+/);

    // TL, TR, BR, BL
    const radii = [0, 0, 0, 0];

    // base radius percent if no r0/r1/r2 present
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

function getDrawingStyle(classes, timeMs, worldY, worldX) {
    const tokens = classes.split(/\s+/);

    let fillColor   = null;
    let strokeColor = null;
    let alpha       = 1.0;
    let borderWidth = 0;

    // handle "invisible" explicitly ?
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

    for (const token of tokens) {
        const cycleMatch = token.match(/^cycle\(([^,]+),([^)]+)\)$/);
        if (cycleMatch) {
            const colorA = COLOR_MAP[cycleMatch[1].trim()];
            const colorB = COLOR_MAP[cycleMatch[2].trim()];
            if (colorA && colorB) {
                const mix = 0.5 + (0.5 * Math.sin((timeMs / 420) + ((worldY + worldX) * 0.3)));
                fillColor = mixColor(colorA, colorB, mix);
            }
            continue;
        }

        const cycleBorderMatch = token.match(/^cycle-b\(([^,]+),([^)]+)\)$/);
        if (cycleBorderMatch) {
            const colorA = COLOR_MAP[cycleBorderMatch[1].trim()];
            const colorB = COLOR_MAP[cycleBorderMatch[2].trim()];
            if (colorA && colorB) {
                const mix = 0.5 + (0.5 * Math.sin((timeMs / 420) + ((worldY + worldX) * 0.3)));
                strokeColor = mixColor(colorA, colorB, mix);
                if (borderWidth === 0) {
                    borderWidth = BORDER_WIDTH_MAP.thin;
                }
            }
            continue;
        }

        if (token === "rainbow") {
            fillColor = rainbowColorAt(timeMs, (worldY + worldX) * 11);
            continue;
        }

        if (token === "rainbow-b") {
            strokeColor = rainbowColorAt(timeMs, (worldY + worldX) * 11);
            if (borderWidth === 0) {
                borderWidth = BORDER_WIDTH_MAP.thin;
            }
            continue;
        }

        if (token === "water") {
            const wave = 0.5 + (0.5 * Math.sin((timeMs / 560) + (worldX * 0.7) + (worldY * 0.35)));
            fillColor = mixColor(COLOR_MAP["blue"], COLOR_MAP["sky-blue"], wave);
            alpha = Math.min(alpha, 0.92);
            continue;
        }
    }

    // If we have a stroke but no fill, that's okay (border-only tile)
    // If we have neither, treat as fully transparent
    if (!fillColor && !strokeColor) {
        return { fillColor: null, strokeColor: null, alpha: 0, borderWidth: 0 };
    }

    return { fillColor, strokeColor, alpha, borderWidth };
}

function hasDynamicTileToken(classes) {
    if (!classes) return false;

    return (
        classes.includes("sparkle") ||
        classes.includes("rainbow") ||
        classes.includes("rainbow-b") ||
        classes.includes("water") ||
        classes.includes("cycle(") ||
        classes.includes("cycle-b(")
    );
}

function mixColor(colorA, colorB, amount) {
    const a = parseRgbColor(colorA);
    const b = parseRgbColor(colorB);
    if (!a || !b) return colorA;

    const t = Math.max(0, Math.min(1, amount));
    const r = Math.round((a.r * (1 - t)) + (b.r * t));
    const g = Math.round((a.g * (1 - t)) + (b.g * t));
    const bCh = Math.round((a.b * (1 - t)) + (b.b * t));

    return `rgb(${r}, ${g}, ${bCh})`;
}

function parseRgbColor(color) {
    const match = color.match(/^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/);
    if (!match) return null;

    return {
        r: Number(match[1]),
        g: Number(match[2]),
        b: Number(match[3]),
    };
}

function rainbowColorAt(timeMs, phase) {
    const hue = ((timeMs / 14) + phase) % 360;
    return hslToRgbString(hue, 0.85, 0.6);
}

function hslToRgbString(h, s, l) {
    const hue = ((h % 360) + 360) % 360;
    const c = (1 - Math.abs((2 * l) - 1)) * s;
    const x = c * (1 - Math.abs(((hue / 60) % 2) - 1));
    const m = l - (c / 2);

    let rPrime = 0;
    let gPrime = 0;
    let bPrime = 0;

    if (hue < 60) {
        rPrime = c; gPrime = x;
    } else if (hue < 120) {
        rPrime = x; gPrime = c;
    } else if (hue < 180) {
        gPrime = c; bPrime = x;
    } else if (hue < 240) {
        gPrime = x; bPrime = c;
    } else if (hue < 300) {
        rPrime = x; bPrime = c;
    } else {
        rPrime = c; bPrime = x;
    }

    const r = Math.round((rPrime + m) * 255);
    const g = Math.round((gPrime + m) * 255);
    const b = Math.round((bPrime + m) * 255);

    return `rgb(${r}, ${g}, ${b})`;
}

function hashFloat(a, b, c, d) {
    const seed =
        (a * 12.9898) +
        (b * 78.233) +
        (c * 37.719) +
        (d * 17.173);
    const value = Math.sin(seed) * 43758.5453123;
    return value - Math.floor(value);
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

////////////////////////////////////////////////////////////////
// Sprites 

const SPRITE_DRAWERS = {
    svgRed:   drawCircle.bind(null, "red"),
    svgGreen: drawCircle.bind(null, "green"),
    svgBlue:  drawCircle.bind(null, "blue"),
};

function drawSpriteCell(id, y, x, classes) {
    const ctx = getCanvasContextByLayerId(id);
    if (!ctx) return;

    const px = Math.round(x * cellSize);
    const py = Math.round(y * cellSize);

    ctx.clearRect(px, py, cellSize, cellSize);

    const tokens = classes.split(/\s+/);

    for (const token of tokens) {
        const fn = SPRITE_DRAWERS[token];
        if (fn) {
            fn(ctx, px, py, cellSize);
        }
    }
}

function drawCircle(colorName, ctx, px, py, cellSize) {
    // positions as fractions of a 22x22 square
    const specByColor = {
        red: {
            cx: 7 / 22,
            cy: 7 / 22,
            r:  7 / 22,
            fill: "rgb(255, 105, 100)",
        },
        green: {
            cx: 7 / 22,
            cy: 14 / 22,
            r:  7 / 22,
            fill: "rgb(105, 255, 100)",
        },
        blue: {
            cx: 14 / 22,
            cy: 14 / 22,
            r:  7 / 22,
            fill: "rgb(15, 100, 255)",
        },
    };

    const spec = specByColor[colorName];
    if (!spec) return;

    const cx = px + cellSize * spec.cx;
    const cy = py + cellSize * spec.cy;
    const r  = cellSize * spec.r;

    ctx.save();
    ctx.fillStyle = spec.fill;
    ctx.globalAlpha = 0.6;
    ctx.beginPath();
    ctx.arc(cx, cy, r, 0, Math.PI * 2);
    ctx.fill();
    ctx.restore();
}
