package main

import (
	"log"
)

// if QuitOnError is true, then panic;
// else go on
func softFailIf(e error) {
	if e != nil {
		log.Printf("[WARNING] %s", e)
	}
}

func hardFailIf(e error) {
	if e != nil {
		log.Printf("[ERROR] %s", e)
		panic(e)
	}
}
