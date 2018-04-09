package utils

// This helper will streamline the error
func Check(e error) {
	if e != nil {
		panic(e)
	}
}