//////////////////////////////////////////////////////////////
// Sound

const sounds = {
    "pop-death": new Audio("/assets/sounds/pop-death.mp3"),
    "money": new Audio("/assets/sounds/money.mp3"),
    "power-up-boost": new Audio("/assets/sounds/power-up-boost.mp3"),
    "power-up-space": new Audio("/assets/sounds/power-up-space.mp3"),
    "teleport": new Audio("/assets/sounds/teleport.mp3"),
    "explosion": new Audio("/assets/sounds/explosion.mp3"),
    "huge-explosion": new Audio("/assets/sounds/huge-explosion.mp3"),
    "wind-swoosh": new Audio("/assets/sounds/wind-swoosh.mp3"),
    "woody-swoosh": new Audio("/assets/sounds/woody-swoosh.mp3"),
    "water-splash": new Audio("/assets/sounds/water-splash.mp3"),
    //success: new Audio("/static/success.mp3"),
};

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
        console.log("Waiting for #sound-trigger to be added...");
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
    console.log("Observing for changes to #sound-trigger");
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