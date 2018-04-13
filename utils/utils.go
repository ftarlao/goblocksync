package utils


// This helper will streamline the error
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func contains(arr []int, el int) bool {
	for _, a := range arr {
		if a == el {
			return true
		}
	}
	return false
}

// Array of common values
func Intersection(arr1 []int, arr2 []int) (intersection []int) {
	intersection = make([]int,0,len(arr1))
	for _, a := range arr1 {
		if contains(arr2,a) {
			intersection = append(intersection, a)
		}
	}
	return intersection
}

func Max(arr []int) (el *int){
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