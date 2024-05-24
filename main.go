package main

import (
	"errors"
	"flag"
	"log"
	"os"
)

func main() {
	listenPort := flag.Int("listen", 7000, "Port to listen on")
	warmupPromptFile := flag.String("warmup-prompt-file", "prompt.txt", "Prompt text file used to warm up llama.cpp")
	llmHost := flag.String("llmhost", "", "hostname:port for llama.cpp server")
	flag.Parse()

	if _, err := os.Stat(*warmupPromptFile); errors.Is(err, os.ErrNotExist) {
		log.Fatal("warmup-prompt-file does not exist: ", *warmupPromptFile)
	}

	wrangler := LlmWrangler{
		ListenPort:       *listenPort,
		warmupPromptFile: *warmupPromptFile,
	}
	wrangler.Init()

	if *llmHost != "" {
		wrangler.RegisterHost(*llmHost)
	}

	wrangler.Start()
}
