// ai.go
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// AI handles interactions with the OpenAI API
type AI struct {
	APIKey string
	Model  string
}

// NewAI creates a new AI instance
func NewAI(apiKey, model string) *AI {
	if apiKey == "" || model == "" {
		panic("OpenAI API key and model must be provided")
	}

	return &AI{
		APIKey: apiKey,
		Model:  model,
	}
}

// Translate returns the Chinese translation and Pinyin pronunciation of the given English sentence
func (ai *AI) Translate(sentence string) (string, string, error) {
	var schema = json.RawMessage([]byte(`{
      "name": "translation",
      "strict": true,
      "schema": {
        "type": "object",
        "properties": {
          "zh": {
            "type": "string"
          },
          "pinyin": {
            "type": "string"
          }
        },
        "required": [
          "zh",
          "pinyin"
        ],
        "additionalProperties": false
      }
    }`))

	var typicalResponse = `{
      "zh": "我下周可能有时间，可以吗？",
      "pinyin": "Wǒ xià zhōu kěnéng yǒu shíjiān, kěyǐ ma?"
    }`

	params := ChatCompletionsParams{
		Messages: []Message{
			{
				Role:    "system",
				Content: "Translate the provided English sentence into Chinese, including pinyin and Chinese characters.",
			},
			{
				Role:    "user",
				Content: "I'll probably have time next week. Is that okay?",
			},
			{
				Role:    "assistant",
				Content: typicalResponse,
			},
			{
				Role:    "user",
				Content: sentence,
			},
		},
		Model: ai.Model,
		ResponseFormat: &ResponseFormat{
			Type:       "json_schema",
			JSONSchema: schema,
		},
	}

	body, err := json.Marshal(params)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Authorization", "Bearer "+ai.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result ChatCompletionsResult
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return "", "", err
	}

	if len(result.Choices) == 0 {
		return "", "", fmt.Errorf("no response from OpenAI API: %s", string(b))
	}

	var translation struct {
		ZH     string `json:"zh"`
		Pinyin string `json:"pinyin"`
	}

	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &translation); err != nil {
		return "", "", err
	}

	if translation.ZH == "" || translation.Pinyin == "" {
		return "", "", errors.New("no translation found")
	}

	return translation.ZH, translation.Pinyin, nil
}
