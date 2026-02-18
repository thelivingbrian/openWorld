package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Collection struct {
	Name             string
	Spaces           map[string]*Space
	Fragments        map[string][]Fragment
	PrototypeSets    map[string][]Prototype
	InteractableSets map[string][]InteractableDescription
	StructureSets    map[string][]Structure
}

func createCollectionDirectories(name string) {
	dirs := []string{"prototypes", "fragments", "spaces", "interactables", "structures"}

	for _, dir := range dirs {
		fullPath := filepath.Join(COLLECTION_PATH, name, dir)
		err := os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
		fmt.Println("Created directory:", fullPath)
	}
}

func (col *Collection) getProtoSelect() PrototypeSelectPage {
	return PrototypeSelectPage{
		PrototypeSets: col.getProtoSets(),
		CurrentSet:    "",
		Prototypes:    nil,
	}
}

func (col *Collection) getProtoSets() []string {
	var setOptions []string
	for key := range col.PrototypeSets {
		setOptions = append(setOptions, key)
	}
	return setOptions
}

func (col *Collection) findPrototypeById(id string) *Prototype {
	for _, set := range col.PrototypeSets {
		for i := range set {
			if set[i].ID == id {
				return &set[i]
			}
		}
	}
	fmt.Println("Invalid Prototype lookup: " + id)
	return nil
}

func (col *Collection) findInteractableById(id string) *InteractableDescription {
	for _, set := range col.InteractableSets {
		for i := range set {
			if set[i].ID == id {
				return &set[i]
			}
		}
	}
	return nil
}

func (col *Collection) savePrototypeSet(setName string) {
	if col == nil {
		fmt.Println("Nil Collection")
		return
	}
	set, ok := col.PrototypeSets[setName]
	if !ok {
		fmt.Println("Empty Prototype Set.")
		return
	}
	save("prototypes", setName, set, col)
}

func (col *Collection) saveFragmentSet(setName string) {
	if col == nil {
		fmt.Println("Nil Collection")
		return
	}
	set, ok := col.Fragments[setName]
	if !ok {
		fmt.Println("Empty Fragment Set.")
		return
	}
	save("fragments", setName, set, col)
}

func (col *Collection) saveInteractableSet(setName string) {
	if col == nil {
		fmt.Println("Nil Collection")
		return
	}
	set, ok := col.InteractableSets[setName]
	if !ok {
		fmt.Println("Empty Interactable Set.")
		return
	}
	save("interactables", setName, set, col)
}

func (col *Collection) saveSpace(name string) {
	if col == nil {
		fmt.Println("Nil Collection")
		return
	}
	space, ok := col.Spaces[name]
	if !ok {
		fmt.Println("Empty Space.")
		return
	}
	save("spaces", name, space, col)
}

func save[T any](directoryName, fileName string, data T, col *Collection) {
	if col == nil {
		panic("Invalid collection")
	}
	outFile := COLLECTION_PATH + col.Name + "/" + directoryName + "/" + fileName + ".json"
	err := writeJsonFile(outFile, data, true)
	if err != nil {
		panic(err)
	}
}
