// Compare the speeds for allocating data to heap and stack and to copying by value
// run with: go test -gcflags '-l' -bench=. -benchmem

package main

type data struct {
	a string
	b string
	x int
}

func NewDataPointerFromStack(d *data) {
	d.a = "Uusi"
	d.b = "data"
}

func NewDataPointerFromHeap() *data {
	d := data{"Uusi", "data", 0}
	return &d
}

func NewDataByValue() data {
	d := data{"Uusi", "data", 0}
	return d
}

func main() {
	caseA()
	caseB()
	caseC()
}

func caseA() {
	var d data
	NewDataPointerFromStack(&d)
	usePointerData(&d)
}

func caseB() {
	d := NewDataPointerFromHeap()
	usePointerData(d)
}

func caseC() {
	d := NewDataByValue()
	usePointerData(&d)
}

func useData(d data) {
	d.a = d.b
}

func usePointerData(d *data) {
	d.a = d.b
}
