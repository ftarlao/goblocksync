package utils

import (
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"bytes"
)

// This helper will streamline the error
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// When array arr SliceContains integer el, returns true, otherwise returns false
func SliceContains(arr []int, el int) bool {
	for _, a := range arr {
		if a == el {
			return true
		}
	}
	return false
}

// Create the Array containing the values in arr1 that are also in arr2, there is no unicity constrain
func SliceIntersection(arr1 []int, arr2 []int) (intersection []int) {
	intersection = make([]int, 0, len(arr1))
	for _, a := range arr1 {
		if SliceContains(arr2, a) {
			intersection = append(intersection, a)
		}
	}
	return intersection
}

// Find max in array, returns nil for empty arrays
func SliceMax(arr []int) (el *int) {
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

//int64 utils
func IntMax(a int64, b int64) int64 {
	if a > b {
		return a
	} else {
		return b
	}
}

//File utils

func IsEOF(err error) bool {
	return err == io.ErrUnexpectedEOF || err == io.EOF
}

// Creates a tmp file in the tmp project folder, the file is randomly generated but can have a periodicity i.e. being made
// by a sequence that repeats.
// periodBytes = 0 means periodicity disabled
// seed is the random seed
func CreateTmpFile(size int64, periodBytes int64, seed int64) (f *os.File, err error) {

	if periodBytes == 0 {
		//Disable periodicity
		periodBytes = size
	}

	tempFileName, err := filepath.Abs("../tmp")
	if err != nil {
		return nil, err
	}
	f, err = ioutil.TempFile(tempFileName, "data")
	if err != nil {
		return nil, err
	}
	data := GeneratePeriodicData(size,periodBytes,seed)
	_, err = f.Write(*data)
	return f, err
}

func GeneratePeriodicData(size int64, periodBytes int64, seed int64) *[]byte {
	source := rand.NewSource(seed)
	rGen := rand.New(source)
	sequence := make([]byte, periodBytes)
	rGen.Read(sequence)
	//repeat sequence!
	data := make([]byte, size)
	for i := int64(0); i < size; i += int64(len(sequence)) {
		copy(data[i:], sequence)
	}
	return &data
}

func CreateTmpRamReader(size int64, periodBytes int64, seed int64) (f *bytes.Reader) {

	if periodBytes == 0 {
		//Disable periodicity
		periodBytes = size
	}

	data := GeneratePeriodicData(size,periodBytes,seed)
	fakeFile := bytes.NewReader(*data)
	return fakeFile
}

//Constants
//Data size bytes

const KB = 1024
const MB = KB * 1024
const GB = MB * 1024
const TB = GB * 1024
