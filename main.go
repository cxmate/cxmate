package main

import (
	"fmt"
	"log"
	"os"
)

const version = "3.0"
const commit = "5d98c0441842029f0a6b7083f975f99a8d159c96"

func main() {
	command := ""
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	switch command {
	case "version":
		fmt.Println("cxMate version", version, "build commit", commit)
	case "config":
		c, err := loadConfig()
		if err != nil {
			log.Fatalln("Error loading cxmate.json:", err)
		}
		fmt.Println("cxmate loaded config:")
		c.Print()
	default:
		m, err := NewMate()
		if err != nil {
			log.Fatalln("Initialization of cxMate failed with error:", err)
		}
		m.serve()
	}
}
