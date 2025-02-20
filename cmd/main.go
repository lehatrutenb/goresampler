package main

import (
	"log"
	"resampler/internal/benchmark"
)

func main() {
	err := benchmark.CreateReadmeAudioTable()
	if err != nil {
		log.Fatal("failed to creade readme audio table", err)
	}
}
