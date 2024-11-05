package main

import (
	"log"
	"os"

	"github.com/aykhans/movier/server/cmd"
	"github.com/aykhans/movier/server/pkg/config"
)

func main() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	config.BaseDir = baseDir

	err = cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
