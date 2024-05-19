package main

import (
	"log"
	"os"
	"strconv"
)

func main() {

	listenPort := 7000
	if os.Getenv("LISTEN") != "" {
		var err error
		listenPort, err = strconv.Atoi(os.Getenv("LISTEN"))
		if err != nil {
			log.Fatal(err)
		}
	}

	wrangler := LlmWrangler{
		ListenPort: listenPort,
	}
	wrangler.Init()
	if os.Getenv("LLMHOST") != "" {
		wrangler.RegisterHost(os.Getenv("LLMHOST"))
	}

	wrangler.Start()
}
