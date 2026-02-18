package main

type Structure struct {
	ID           string        `json:"id"`
	FragmentIds  [][]string    `json:"fragmentIds"`
	GroundConfig *GroundConfig `json:"groundConfig,omitempty"`
}

type GroundConfig struct {
	Name     string  `json:"name"`
	Span     int     `json:"span"`
	Color1   string  `json:"color1"`
	Color2   string  `json:"color2"`
	Fuzz     float64 `json:"fuzz"`
	Strategy string  `json:"strategy"`
}

func removeStructuresById(structures []Structure, id string) []Structure {
	out := make([]Structure, 0)
	for i := range structures {
		if structures[i].ID != id {
			out = append(out, structures[i])
		}
	}
	return out
}

func getStructureById(structures []Structure, id string) (Structure, bool) {
	for i := range structures {
		if structures[i].ID == id {
			return structures[i], true
		}
	}
	return Structure{}, false
}
