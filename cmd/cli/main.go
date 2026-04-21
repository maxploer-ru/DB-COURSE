package main

import (
	"flag"
	"log"
)

func main() {
	apiURL := flag.String("api-url", "http://localhost:8080", "Базовый URL ZVideo API")
	flag.Parse()
	SetBaseURL(*apiURL)
	log.SetFlags(0)
	runMainMenu()
}
