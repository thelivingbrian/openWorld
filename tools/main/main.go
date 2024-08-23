package main

import (
	"fmt"
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.ParseGlob("templates/*.tmpl.html"))

func main() {
	fmt.Println("Attempting to start server...")
	c := populateFromJson()

	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))

	http.HandleFunc("/collections", c.collectionsHandler)
	http.HandleFunc("/spaces", c.spacesHandler)
	http.HandleFunc("/spaces/new", c.newSpaceHandler)
	http.HandleFunc("/space", c.spaceHandler)
	http.HandleFunc("/space/map", c.spaceMapHandler)
	http.HandleFunc("/areas", c.areasHandler)
	http.HandleFunc("/areas/new", c.newAreaHandler)
	http.HandleFunc("/area", c.areaHandler)
	http.HandleFunc("/area/details", c.areaDetailsHandler)
	http.HandleFunc("/area/display", c.areaDisplayHandler)
	http.HandleFunc("/area/neighbors", c.areaNeighborsHandler)
	http.HandleFunc("/area/blueprint", c.areaBlueprintHandler)
	http.HandleFunc("/area/blueprint/instruction", c.blueprintInstructionHandler)
	http.HandleFunc("/area/blueprint/instruction/highlight", c.blueprintInstructionHighlightHandler)
	http.HandleFunc("/area/blueprint/instructions/order", c.instructionOrderHandler)
	http.HandleFunc("/fragments", c.fragmentsHandler)
	http.HandleFunc("/fragments/new", c.fragmentsNewHandler)
	http.HandleFunc("/fragment", c.fragmentHandler)
	http.HandleFunc("/fragment/new", c.fragmentNewHandler)
	http.HandleFunc("/prototypes", c.prototypesHandler)
	http.HandleFunc("/prototypes/new", c.prototypesNewHandler)
	http.HandleFunc("/prototype", c.prototypeHandler)
	http.HandleFunc("/prototype/new", c.prototypeNewHandler)
	http.HandleFunc("/prototype/example", examplePrototype)
	http.HandleFunc("/grid/edit", c.gridEditHandler)
	http.HandleFunc("/grid/click/area", c.gridClickAreaHandler)
	http.HandleFunc("/grid/click/fragment", c.gridClickFragmentHandler)
	http.HandleFunc("/images/", c.imageHandler)

	http.HandleFunc("/materialPage", c.getMaterialPage)
	//http.HandleFunc("/exampleMaterial", exampleMaterial) // Probably unused
	http.HandleFunc("/getEditColor", c.getEditColor)
	http.HandleFunc("/editColor", c.editColor)
	http.HandleFunc("/getNewColor", getNewColor)
	http.HandleFunc("/newColor", c.newColor)
	http.HandleFunc("/exampleSquare", exampleSquare)
	http.HandleFunc("/outputIngredients", c.outputIngredients)

	http.HandleFunc("/selectFixture", c.selectFixture)

	http.HandleFunc("/editTransports", c.getEditTransports)
	http.HandleFunc("/editTransport", c.editTransport)
	http.HandleFunc("/newTransport", c.newTransport)
	http.HandleFunc("/dupeTransport", c.dupeTransport)
	http.HandleFunc("/deleteTransport", c.deleteTransport)

	http.HandleFunc("/deploy", c.deploy)
	http.HandleFunc("/compile", c.compile)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmpl.ExecuteTemplate(w, "home", c.Collections)
	})

	err := http.ListenAndServe(":4444", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}

type Coordinate struct {
	x int
	y int
}

type Stack struct {
	elements []Coordinate
}

func (s *Stack) Push(element Coordinate) {
	s.elements = append(s.elements, element)
}

func (s *Stack) Pop() Coordinate {
	if len(s.elements) == 0 {
		fmt.Println("Stack is empty!")
		return Coordinate{-1, -1}
	}
	element := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return element
}

func (s *Stack) Peek() Coordinate {
	if len(s.elements) == 0 {
		fmt.Println("Stack is empty!")
		return Coordinate{-1, -1}
	}
	return s.elements[len(s.elements)-1]
}

func (s *Stack) IsEmpty() bool {
	return len(s.elements) == 0
}

func findValueInMatrix(matrix [][]int, x, y, target int) bool {
	var seen = make([][]bool, len(matrix))
	for i := range seen {
		seen[i] = make([]bool, len(matrix[i]))
		for j := range seen[i] {
			seen[i][j] = false
		}
	}

	printSeenMatrix(seen)

	stack := Stack{}
	stack.Push(Coordinate{x, y})
	var found = false
	for !found {
		if !stack.IsEmpty() {
			coord := stack.Pop()
			if coord.x >= 0 && coord.y >= 0 && coord.y < len(matrix) && coord.y < len(matrix[0]) && matrix[coord.x][coord.y] == target {
				seen[coord.x][coord.y] = true
				return true
			}

			if validCoordinate(coord.x-1, coord.y, seen) && !seen[coord.x-1][coord.y] {
				if matrix[coord.x-1][coord.y] == target {
					return true
				} else {
					//seen[coord.x-1][coord.y] = true
					printSeenMatrix(seen)
					stack.Push(Coordinate{x - 1, y})
				}
			}

			if validCoordinate(coord.x+1, coord.y, seen) && !seen[coord.x+1][coord.y] {
				if matrix[coord.x+1][coord.y] == target {
					return true
				} else {
					//seen[coord.x+1][coord.y] = true
					printSeenMatrix(seen)
					stack.Push(Coordinate{x + 1, y})
				}
			}

			if validCoordinate(coord.x, coord.y-1, seen) && !seen[coord.x][coord.y-1] {
				if matrix[coord.x][coord.y-1] == target {
					return true
				} else {
					//seen[coord.x][coord.y-1] = true
					printSeenMatrix(seen)
					stack.Push(Coordinate{coord.x, coord.y - 1})

				}
			}
			if validCoordinate(coord.x, coord.y+1, seen) && !seen[coord.x][coord.y+1] {
				if matrix[coord.x][coord.y+1] == target {
					return true
				} else {
					//seen[coord.x][coord.y+1] = true
					printSeenMatrix(seen)
					stack.Push(Coordinate{coord.x, coord.y + 1})
				}
			}
		}
	}

	return false
}

func printSeenMatrix(seen [][]bool) {
	fmt.Println("--------------------------")
	fmt.Println(seen)
	fmt.Println("--------------------------")
}

func validCoordinate(x int, y int, tiles [][]bool) bool {
	if x < 0 || x >= len(tiles) {
		return false
	}
	if y < 0 || y >= len(tiles[x]) {
		return false
	}
	return true
}
