package main

import "fmt"

type T struct {
	Name string
	Age int64
	T2
}

type T2 []int

func main () {
	t := new(T)

	fmt.Println(t.T2 == nil)
}


