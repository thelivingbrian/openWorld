package main

import (
	"fmt"
	"testing"
)

func BenchmarkInsert(b *testing.B) {
	fmt.Println("Hi")
}

func BenchmarkUpdate(b *testing.B) {
	fmt.Println("Hi")
}
