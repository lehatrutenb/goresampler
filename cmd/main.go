package main

import (
	"log"

	"github.com/lehatrutenb/goresampler/internal/benchmark"
)

func main() {
	err := benchmark.CreateReadmeAudioTable()
	if err != nil {
		log.Fatal("failed to creade readme audio table", err)
	}
}
