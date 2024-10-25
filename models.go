// models.go
package main

import "encoding/json"

// Flashcard represents a single card in the deck
type Flashcard struct {
	ID      int    `json:"id"`
	English string `json:"en"`
	Chinese string `json:"zh"`
	Pinyin  string `json:"pinyin"`
}

// Message represents a message to or from the AI
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseFormat represents the response format for the AI API
type ResponseFormat struct {
	Type       string          `json:"type"`
	JSONSchema json.RawMessage `json:"json_schema"`
}

// ChatCompletionsParams represents the parameters for the chat completions API
type ChatCompletionsParams struct {
	Messages            []Message       `json:"messages"`
	Model               string          `json:"model"`
	MaxCompletionTokens *int            `json:"max_completion_tokens,omitempty"`
	Temperature         *float64        `json:"temperature,omitempty"`
	ResponseFormat      *ResponseFormat `json:"response_format,omitempty"`
}

// ChatCompletionsResult represents the result from the chat completions API
type ChatCompletionsResult struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}
