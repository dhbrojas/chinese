// main.go
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	apiKey := flag.String("api-key", "", "OpenAI API key (required)")
	filePath := flag.String("file", "flashcards.jsonl", "Path to flashcards file")
	model := flag.String("model", "gpt-4o-mini", "OpenAI model to use")
	flag.Parse()

	if *apiKey == "" {
		*apiKey = os.Getenv("OPENAI_API_KEY")
		if *apiKey == "" {
			fmt.Println("Please provide an API key")
			os.Exit(1)
		}
	}

	app := NewApp(*apiKey, *model)

	// Load the deck
	if err := app.LoadDeck(*filePath); err != nil {
		fmt.Printf("Error loading deck: %v\n", err)
		os.Exit(1)
	}

	app.SetupUI()
	app.Application.SetInputCapture(app.HandleInput)

	if err := app.Application.SetRoot(app.MainView, true).Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
