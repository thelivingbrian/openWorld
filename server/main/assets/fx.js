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

var topLeftY = 0
var topLeftX = 0
const height = 17 // Camera height and Camera Width 
const width = 17

var cells = undefined

function setGrid(y, x) {
  cells = undefined 
  topLeftY = Number(y)
  topLeftX = Number(x)
}

/**
 * Shift the visible grid by (dy, dx) cells.
 *  •  dy  < 0  →  move view *up*    (copy rows upward)
 *  •  dy  > 0  →  move view *down*  (copy rows downward)
 *  •  dx  < 0  →  move view *left*  (copy cols leftward)
 *  •  dx  > 0  →  move view *right* (copy cols rightward)
 *
 */
function shiftGrid(dy, dx) {
  if (dy === 0 && dx === 0) return;

  if (!cells) {
    cells = Array.from({ length: height }, (_, r) =>
      Array.from({ length: width },  (_, c) =>
        document.getElementById(`c${r}-${c}`)
      )
    );
  }

  topLeftX -= dx;
  topLeftY -= dy;

  /* ----------------------------------------------------------
   *  Determine safe traversal order so we never overwrite
   *  a source tile before we’ve copied out its contents.
   *     – When shifting toward +y or +x we walk bottom‑right→top‑left
   *     – When shifting toward –y or –x we walk top‑left→bottom‑right
   * -------------------------------------------------------- */
  const rowStep   = dy > 0 ? -1 : 1;
  const colStep   = dx > 0 ? -1 : 1;

  const rowStart  = dy > 0 ? height - 1 - dy : (dy < 0 ? -dy : 0);
  const rowEnd    = dy > 0 ? -1              : height;      // exclusive
  const colStart  = dx > 0 ? width  - 1 - dx : (dx < 0 ? -dx : 0);
  const colEnd    = dx > 0 ? -1              : width;       // exclusive

  for (let r = rowStart; r !== rowEnd; r += rowStep) {
    for (let c = colStart; c !== colEnd; c += colStep) {
      const here  = cells[r][c];
      const there = cells[r + dy][c + dx];
      shiftChildClasses(here, there);
    }
  }
}

function shiftChildClasses(a, b) {
  for (let i = 0; i < a.children.length; i++) {
    b.children[i].className = a.children[i].className;
  }
}

/*
function shiftGrid(dir) {
  const { dr, dc, rowStart, rowEnd, colStart, colEnd } = {
    left:  { dr: 0, dc:-1, rowStart: 0,         rowEnd: height, colStart: 1,        colEnd: width   },
    right: { dr: 0, dc: 1, rowStart: 0,         rowEnd: height, colStart: width-2,  colEnd:-1 },
    up:    { dr:-1, dc: 0, rowStart: 1,         rowEnd: height, colStart: 0,        colEnd: width   },
    down:  { dr: 1, dc: 0, rowStart: height-2,  rowEnd:-1,      colStart: 0,        colEnd: width   }
  }[dir];

  if (!cells) {
    cells = Array.from({ length: height }, (_, r) =>
                Array.from({ length: width }, (_, c) =>
                    document.getElementById(`c${r}-${c}`)
                )
            );
  }

  topLeftX -= dc
  topLeftY -= dr


  for (let r = rowStart; r !== rowEnd; r += Math.sign(rowEnd - rowStart || 1)) {
    for (let c = colStart; c !== colEnd; c += Math.sign(colEnd - colStart || 1)) {
      const here = cells[r][c];
      const there = cells[r + dr][c + dc];
      shiftChildClasses(here, there);
    }
  }
}

function shiftChildClasses(a, b) {
  for (let i = 0; i < a.children.length; i++) {
    b.children[i].className = a.children[i].className;
  }
}
*/