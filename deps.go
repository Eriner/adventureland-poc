package main

import (
	"log"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	if err := playwright.Install(); err != nil {
		log.Fatal(err)
	}
}
