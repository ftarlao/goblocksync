package test

import (
	"testing"
	"github.com/ftarlao/goblocksync/utils"
	"reflect"
)

func TestUnitSliceContains(t *testing.T){
	t.Log("***SliceContains Test***")
	var slc1 = []int{0,2,3,4}
	if !utils.SliceContains(slc1,2) || utils.SliceContains(slc1,5) {
		t.Error("Test failed")
	} else {
		t.Log("Test ok")
	}
}

func TestUnitSliceIntersection(t *testing.T){
	t.Log("***SliceContains Test***")
	var slc1 = []int{0,2,3,4}
	var slc2 = []int{3,4}
	var slc3 = []int{0}
	var slc4 = []int{}

	success := true
	success =  success && testSliceIntersection(slc1,slc2,[]int{3,4},t)
	success =  success && testSliceIntersection(slc2,slc1,[]int{3,4},t)
	success =  success && testSliceIntersection(slc1,slc3,[]int{0},t)
	success =  success && testSliceIntersection(slc3,slc1,[]int{0},t)
	success =  success && testSliceIntersection(slc1,slc4,[]int{},t)
	success =  success && testSliceIntersection(slc4,slc1,[]int{},t)
	if success {
		t.Log("Test ok")
	}
}

func testSliceIntersection(a []int, b []int, c []int, t *testing.T) bool{
	if irs :=  utils.SliceIntersection(a,b); !reflect.DeepEqual(irs, c) {
		t.Error("Test failed with a = ",a," b = ",b," intersection = ",c)
		return false
	}
	return true
}