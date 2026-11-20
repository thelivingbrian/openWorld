//////////////////////////////////////////////////////////////
// Sound

const sounds = {
    "pop-death": new Audio("/assets/sounds/pop-death.mp3"),
    "money": new Audio("/assets/sounds/money.mp3"),
    "power-up-boost": new Audio("/assets/sounds/power-up-boost.mp3"),
    "power-up-space": new Audio("/assets/sounds/power-up-space.mp3"),
    "teleport": new Audio("/assets/sounds/teleport.mp3"),
    "explosion": new Audio("/assets/sounds/explosion.mp3"),
    "huge-explosion": new Audio("/assets/sounds/huge-explosion-in-distance.mp3"),
    "wind-swoosh": new Audio("/assets/sounds/wind-swoosh.mp3"),
    "woody-swoosh": new Audio("/assets/sounds/woody-swoosh.mp3"),
    "water-splash": new Audio("/assets/sounds/water-splash.mp3"),
    "clink": new Audio("/assets/sounds/clink.mp3"),
};

// Mixing
sounds["explosion"].volume = 0.5;
sounds["power-up-space"].volume = 0.3;
sounds["water-splash"].volume = 0.7;

function playSound(soundName) {
    const sound = sounds[soundName];
    if (sound) {
        sound.currentTime = 0; // Reset if already playing
        sound.play().catch(err => console.log("Autoplay blocked:", err));
    } else {
        console.log(`Sound "${soundName}" not found.`);
    }
}

function observeSoundTrigger() {
    const soundTrigger = document.getElementById("sound-trigger");
    if (!soundTrigger) {
        setTimeout(observeSoundTrigger, 500);
        return;
    }

    const observer = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
            mutation.addedNodes.forEach((node) => {
                if (node.id === "sound") {
                    const soundName = node.innerText.trim();
                    if (soundName) {
                        playSound(soundName);
                    }
                }
            });
        });
    });
    observer.observe(soundTrigger, { childList: true, subtree: true });
}

observeSoundTrigger()


////////////////////////////////////////////////////////////////
//  Video

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function flashBg(color){
    document.body.className=color
    await sleep(10)
    document.body.className="night"
}


////////////////////////////////////////////////////////////////
//  Camera

var topLeftY = 0
var topLeftX = 0
let height = 16 // Negotiated with the server; square world tiles.
let width = 16
let stageHeight = 16, stageWidth = 16;

function setGrid(y, x) {
    resetStageBuffers();
    topLeftY = Number(y);
    topLeftX = Number(x);
    scheduleStageDraw();
}

function shiftGrid(dy, dx) {
    if (dy === 0 && dx === 0) return;

    topLeftY -= dy;
    topLeftX -= dx;

    scheduleStageDraw();
}

///////////////////////////////////////////////////////////////
//  Mobile Controls

function addRepeater(btn) {
    if (btn.dataset.repeater) return;
    btn.dataset.repeater = 'true';
    let timer, interval, pointer;
    const lifecycle = new AbortController();
    const options = { signal: lifecycle.signal };
    const stop = () => {
        clearTimeout(timer); clearInterval(interval); pointer = undefined;
        btn.classList.remove('pressed');
    };
    const fire = () => {
        if (!btn.isConnected || document.hidden || document.querySelector('.modal_bg')) { stop(); return; }
        btn.dispatchEvent(new CustomEvent('fire', { bubbles: true }));
    };
    btn.addEventListener('pointerdown', e => {
        if (pointer !== undefined) return;
        e.preventDefault(); stop(); pointer = e.pointerId;
        btn.setPointerCapture(e.pointerId); btn.classList.add('pressed'); fire();
        timer = setTimeout(() => { interval = setInterval(fire, 90); }, 220);
    });
    ['pointerup', 'pointercancel', 'lostpointercapture'].forEach(t => btn.addEventListener(t, stop));
    window.addEventListener('blur', stop, options);
    document.addEventListener('visibilitychange', stop, options);
    document.addEventListener('htmx:beforeSwap', stop, options);
    document.addEventListener('htmx:beforeCleanupElement', e => {
        if (e.target === btn || e.target.contains(btn)) { stop(); lifecycle.abort(); }
    }, options);
}
function enableKeyRepeat() {
    document.querySelectorAll('#dpad button[data-repeat], #dpad-shift button[data-repeat]').forEach(addRepeater);
}
