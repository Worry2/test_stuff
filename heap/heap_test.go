package main

import (
	"testing"
)

func BenchmarkPointerStack(b *testing.B) {
	for n := 0; n < b.N; n++ {
		caseA()
	}
}

func BenchmarkPointerHeap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		caseB()
	}
}

func BenchmarkValue(b *testing.B) {
	for n := 0; n < b.N; n++ {
		caseC()
	}
}
