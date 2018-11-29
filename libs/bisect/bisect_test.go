package bisect

import (
	"fmt"
	"testing"
)

func generateList() ListInt {
	tl := make(ListInt, 15)
	for i := 0; i <= 10; i++ {
		tl[i] = i
	}
	tl[11] = 23
	tl[12] = 33
	tl[13] = 34
	tl[14] = 99
	return tl
}

func TestBisectInsertLeft(t *testing.T) {

	tl := generateList()
	fmt.Println(tl)

	list, err := InsertLeft(tl, 3)
	if err != nil {
		fmt.Println(err)
		return
	}

	if v, ok := list.(ListInt); ok {
		fmt.Println(v)
	}
}

func TestBisectSearchLeft(t *testing.T) {
	tl := generateList()
	list, _ := InsertLeft(tl, 7)
	tl = list.(ListInt)
	list, _ = InsertLeft(tl, 7)
	tl = list.(ListInt)
	list, err := InsertLeft(tl, 7)
	if err != nil {
		fmt.Println(err)
	}

	tl = list.(ListInt)
	fmt.Println(SearchInsertPostLeft(tl, 7))
	fmt.Println(tl)
}

func TestBisectInsertRight(t *testing.T) {
	tl := generateList()
	list, err := InsertRight(tl, 11)
	if err != nil {
		fmt.Println(err)
	}

	if v, ok := list.(ListInt); ok {
		fmt.Println(v)
	}
}

func TestBisectSearchRight(t *testing.T) {
	tl := generateList()
	list, _ := InsertRight(tl, 7)
	tl = list.(ListInt)

	list, _ = InsertRight(tl, 7)
	tl = list.(ListInt)

	list, err := InsertRight(tl, 7)
	if err != nil {
		fmt.Println(err)
	}

	tl = list.(ListInt)

	fmt.Println(SearchInsertPosRight(tl, 7))
	fmt.Println(tl)
}
