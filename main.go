package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Flashcard represents a single card in the deck
type Flashcard struct {
	ID     int    `json:"id"`
	EN     string `json:"en"`
	ZH     string `json:"zh"`
	Pinyin string `json:"pinyin"`
}

// App holds the application state
type App struct {
	ai           *AI
	deck         []Flashcard
	currentCard  int
	revealed     bool
	app          *tview.Application
	mainView     *tview.Flex
	cardView     *tview.TextView
	newCardModal *tview.Modal
	inputField   *tview.InputField
}

func newApp(apiKey, model string) *App {
	return &App{
		ai:          newAI(apiKey, model),
		deck:        make([]Flashcard, 0),
		currentCard: 0,
		revealed:    false,
		app:         tview.NewApplication(),
	}
}

// loadDeck loads flashcards from a JSONL file
func (a *App) loadDeck(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var card Flashcard
		if err := decoder.Decode(&card); err != nil {
			return err
		}
		a.deck = append(a.deck, card)
	}
	return nil
}

// updateCardView updates the display of the current card
func (a *App) updateCardView() {
	if len(a.deck) == 0 {
		a.cardView.SetText("No cards in deck!")
		return
	}

	card := a.deck[a.currentCard]
	var content strings.Builder
	content.WriteString("\n\n\n") // Add some padding at the top
	content.WriteString(fmt.Sprintf("Card %d/%d\n\n", a.currentCard+1, len(a.deck)))
	content.WriteString("English:\n")
	content.WriteString(card.EN + "\n\n")

	if a.revealed {
		content.WriteString("Chinese:\n")
		content.WriteString(card.ZH + "\n\n")
		content.WriteString("Pinyin:\n")
		content.WriteString(card.Pinyin + "\n")
	}

	content.WriteString("\n─────────────────────────\n")
	content.WriteString("\nControls:\n")
	content.WriteString("→: Reveal/Next Card  |  n: New Card  |  q: Quit")

	a.cardView.SetText(content.String())
}

// setupUI initializes the user interface
func (a *App) setupUI() {
	// Create the main card view
	a.cardView = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Create the input field for new cards
	a.inputField = tview.NewInputField().
		SetLabel("English: ").
		SetFieldWidth(50)

	// Create the modal for new cards
	a.newCardModal = tview.NewModal().
		SetText("Enter new flashcard details").
		AddButtons([]string{"Save", "Cancel"})

	// Set up the main view
	a.mainView = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(a.cardView, 0, 2, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	// Style the card view
	a.cardView.SetBorder(true).
		SetTitle(" Chinese Learning Cards ").
		SetTitleAlign(tview.AlignCenter)

	a.updateCardView()
}

// handleInput processes keyboard input
func (a *App) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if _, ok := a.app.GetFocus().(*tview.Form); ok {
		return event
	}
	if _, ok := a.app.GetFocus().(*tview.InputField); ok {
		return event
	}

	switch event.Key() {
	case tcell.KeyRight:
		if !a.revealed {
			a.revealed = true
		} else {
			a.revealed = false
			a.currentCard = (a.currentCard + 1) % len(a.deck)
		}
		a.updateCardView()
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			a.app.Stop()
		case 'n':
			a.showNewCardDialog()
		}
	}
	return event
}

// saveNewCard handles saving a new flashcard
func (a *App) saveNewCard(englishText string) {
	zh, pinyin, err := a.ai.Translate(englishText)
	if err != nil {
		a.app.Stop()
		fmt.Println("Error translating text:", err)
		return
	}

	// Here you would make the OpenAI API call
	// For now, we'll just add a placeholder card
	newCard := Flashcard{
		ID:     len(a.deck) + 1,
		EN:     englishText,
		ZH:     zh,
		Pinyin: pinyin,
	}
	a.deck = append(a.deck, newCard)
	a.app.SetRoot(a.mainView, true)
	a.updateCardView()
}

// showNewCardDialog displays the new card input dialog
func (a *App) showNewCardDialog() {
	var englishInput *tview.InputField

	form := tview.NewForm()
	englishInput = tview.NewInputField().
		SetLabel("English").
		SetFieldWidth(50)

	form.AddFormItem(englishInput)
	form.AddButton("Save", func() {
		a.saveNewCard(englishInput.GetText())
	})
	form.AddButton("Cancel", func() {
		a.app.SetRoot(a.mainView, true)
	})

	form.SetBorder(true).
		SetTitle(" Add New Card ").
		SetTitleAlign(tview.AlignCenter)

	formFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 0, 1, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	a.app.SetRoot(formFlex, true)
}

type AI struct {
	key   string
	model string
}

func newAI(key, model string) *AI {
	if key == "" || model == "" {
		panic("OpenAI key and model must be provided")
	}

	return &AI{
		key:   key,
		model: model,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormat struct {
	Type       string          `json:"type"`
	JSONSchema json.RawMessage `json:"json_schema"`
}

type ChatCompletionsParams struct {
	Messages            []Message       `json:"messages"`
	Model               string          `json:"model"`
	MaxCompletionTokens *int            `json:"max_completion_tokens,omitempty"`
	Temperature         *float64        `json:"temperature,omitempty"`
	ResponseFormat      *ResponseFormat `json:"response_format,omitempty"`
}

type ChatCompletionsResult struct {
	Choices []struct {
		Message Message `json:"message"`
	}
}

// Returns the Chinese translation and Pinyin pronunciation of the given English sentence
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
		Model: ai.model,
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

	req.Header.Set("Authorization", "Bearer "+ai.key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result ChatCompletionsResult
	b, err := io.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &result); err != nil {
		return "", "", err
	}

	if len(result.Choices) == 0 {
		return "", "", fmt.Errorf("no response from OpenAI API: %s: %s", string(b), string(body))
	}

	var flashcard Flashcard

	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &flashcard); err != nil {
		return "", "", err
	}

	if flashcard.ZH == "" || flashcard.Pinyin == "" {
		return "", "", errors.New("no translation found")
	}

	return flashcard.ZH, flashcard.Pinyin, nil
}

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

	app := newApp(*apiKey, *model)

	// Load the deck (you'll need to provide the path to your JSONL file)
	if err := app.loadDeck(*filePath); err != nil {
		fmt.Printf("Error loading deck: %v\n", err)
		os.Exit(1)
	}

	app.setupUI()
	app.app.SetInputCapture(app.handleInput)

	if err := app.app.SetRoot(app.mainView, true).EnableMouse(true).Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
