package main

type Fragment struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	SetName   string     `json:"setName"`
	Blueprint *Blueprint `json:"blueprint"`
}
