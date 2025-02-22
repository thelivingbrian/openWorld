package main

import (
	"math/rand/v2"
	"strconv"
)

var honorifics = []string{
	"",
	"Mr ",
	"Ms ",
	"King ",
	"Queen ",
	"Duke ",
	"Duchess ",
	"Prince ",
	"Princess ",
	"Sir ",
	"Madam ",
	"Lord ",
	"Lady ",
	"Baron ",
	"Baroness ",
	"Earl ",
	"Count ",
	"Countess ",
	"Emperor ",
	"Empress ",
	"Archduke ",
	"Archduchess ",
	"Sultan ",
	"Sheikh ",
	"The ",
	"One ",
	"Some ",
	"A ",
	"Not a ",
	"Just ",
	"Old ",
	"Once ",
	"Still ",
	"Another ",
	"Big ",
	"Bold ",
	"Former ",
	"Ex ",
	"Top ",
	"Most ",
	"Least ",
	"Overly ",
	"Esteemed ",
	"Cryptic ",
	"Crazy ",
	"Cpt ",
	"Major ",
	"General ",
	"Dr ",
	"Judge ",
}

var adjectives = []string{
	"Big-",
	"Brave-",
	"Calm-",
	"Clever-",
	"Cold-",
	"Crisp-",
	"Crunchy-",
	"Cursed-",
	"Cute-",
	"Dark-",
	"Drama-",
	"Eager-",
	"Faint-",
	"Fair-",
	"Fast-",
	"Firm-",
	"Fresh-",
	"Gentle-",
	"Glad-",
	"Grim-",
	"Grand-",
	"Happy-",
	"Hidden-",
	"Keen-",
	"Kind-",
	"Light-",
	"Lithe-",
	"Lucky-",
	"Mild-",
	"Neat-",
	"Old-",
	"OK-",
	"Pious-",
	"Proud-",
	"Pretty-",
	"Quick-",
	"Rare-",
	"Rich-",
	"Rude-",
	"Salty-",
	"Seafaring-",
	"Scheming-",
	"Sharp-",
	"Slippery-",
	"Special-",
	"Small-",
	"Strong-",
	"Sweet-",
	"Tame-",
	"Ugly-",
}

var numerals = []string{
	"",
	" II",
	" III",
	" IV",
	" V",
	" VI",
	" VII",
	" VIII",
	" IX",
	" X",
	" XI",
	" XII",
	" PHD",
}

func generateRandomName() string {
	a := rand.IntN(len(honorifics))
	b := rand.IntN(len(adjectives))
	return honorifics[a] + adjectives[b] + "Bloop"
}

func (db *DB) UniqueName() string {
	base := generateRandomName()
	for i := 0; i < len(numerals); i++ {
		candidate := base + numerals[i]
		if !db.foundUsername(candidate) {
			return candidate
		}
	}
	salt := strconv.Itoa(rand.IntN(10000))
	return base + salt // May not be unique but oh well
}
