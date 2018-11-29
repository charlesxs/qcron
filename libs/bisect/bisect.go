package bisect

import (
	"errors"
)

type Bisector interface {
	Len() int
	Less(v interface{}, index int) bool
	LessEqual(v interface{}, index int) bool
	Insert(v interface{}, index int) (interface{}, error)
}

func SearchInsertPosRight(list Bisector, x interface{}) int  {
	var lo, hi, mid int
	hi, lo = list.Len(), 0

	for lo < hi {
		mid = int((lo + hi) / 2)
		if list.Less(x, mid) {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

func InsertRight(list Bisector, x interface{}) (interface{}, error) {
	pos := SearchInsertPosRight(list, x)
	return list.Insert(x, pos)
}

func SearchInsertPostLeft(list Bisector, x interface{}) int {
	var lo, hi, mid int
	hi, lo = list.Len(), 0

	for lo < hi {
		mid = int((lo + hi) / 2)
		if list.LessEqual(x, mid) {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}


func InsertLeft(list Bisector, x interface{}) (interface{}, error) {
	pos := SearchInsertPostLeft(list, x)
	return list.Insert(x, pos)
}


type ListInt []int

func (l ListInt) Len() int {
	return len(l)
}

func (l ListInt) Less(x interface{}, index int) bool {
	v, ok := x.(int)
	if !ok {
		return false
	}

	if v < l[index] {
		return true
	}
	return false
}

func (l ListInt) LessEqual(x interface{}, index int) bool {
	v, ok := x.(int)
	if !ok {
		return false
	}

	if v <= l[index] {
		return true
	}
	return false
}

func (l ListInt) Insert(x interface{}, index int) (interface{}, error) {
	v, ok := x.(int)
	if !ok {
		return nil, errors.New("must be int type")
	}

	var newList = make(ListInt, len(l) + 1)
	copy(newList[:index], l[:index])
	newList[index] = v
	copy(newList[index+1:], l[index:])
	return newList, nil
}
