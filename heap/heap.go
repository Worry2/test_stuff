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

}

func useData(d data) {
	d.a = d.b
}

func usePointerData(d *data) {
	d.a = d.b
}
