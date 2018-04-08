package controller

import (
	"os"
	"fmt"
	"goblocksync/utils"
	"goblocksync/data"
)

type Source struct {
	Config data.Configuration
	currentPosition int64
	sourceFile     *os.File
}

func (s Source) Start(){
	// You'll often want more control over how and what
	// parts of a file are read. For these tasks, start
	// by `Open`ing a file to obtain an `os.File` value.
	f, err := os.Open(s.Config.SourceFileName)
	utils.Check(err)

	_, err = f.Seek(s.currentPosition, 0)
	utils.Check(err)

	// Read some bytes from the beginning of the file.
	// Allow up to 5 to be read but also note how many
	// actually were read.
	b1 := make([]byte, 2048)
	n1, err := f.Read(b1)
	utils.Check(err)
	fmt.Printf("%d bytes: %s\n", n1, string(b1))
}
