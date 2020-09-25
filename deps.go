package main

import (
	"log"

	"github.com/eriner/playwright-go"
)

func main() {
	if err := playwright.Install(); err != nil {
		log.Fatal(err)
	}
}
