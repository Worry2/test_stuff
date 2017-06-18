package main

import (
	"testing"
)

func BenchmarkPointerStack(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var d data
		NewDataPointerFromStack(&d)
		usePointerData(&d)
	}
}

func BenchmarkPointerHeap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		d := NewDataPointerFromHeap()
		usePointerData(d)
	}
}

func BenchmarkValue(b *testing.B) {
	for n := 0; n < b.N; n++ {
		d := NewDataByValue()
		usePointerData(&d)
	}
}
