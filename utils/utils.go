package utils

import (
	"encoding/gob"
	"io"
)

// This helper will streamline the error
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// When array arr contains integer el, returns true, otherwise returns false
func contains(arr []int, el int) bool {
	for _, a := range arr {
		if a == el {
			return true
		}
	}
	return false
}

// Create the Array containing the values in arr1 that are also in arr2, there is no unicity constrain
func Intersection(arr1 []int, arr2 []int) (intersection []int) {
	intersection = make([]int, 0, len(arr1))
	for _, a := range arr1 {
		if contains(arr2, a) {
			intersection = append(intersection, a)
		}
	}
	return intersection
}

// Find max in array, returns nil for empty arrays
func Max(arr []int) (el *int) {
	if len(arr) == 0 {
		return nil
	}
	maxim := arr[0]
	for _, a := range arr {
		if a > maxim {
			maxim = a
		}
	}
	return &maxim
}

func DoNothing(a interface{}) {

}
