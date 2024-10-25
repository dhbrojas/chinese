// app.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App holds the application state
type App struct {
	AI             *AI
	Deck           []Flashcard
	CurrentCardIdx int
	Revealed       bool
	Application    *tview.Application
	MainView       *tview.Flex
	CardView       *tview.TextView
	FlashcardsFile string
}

// NewApp creates a new application instance
func NewApp(apiKey, model string) *App {
	return &App{
		AI:             NewAI(apiKey, model),
		Deck:           make([]Flashcard, 0),
		CurrentCardIdx: 0,
		Revealed:       false,
		Application:    tview.NewApplication(),
	}
}

// LoadDeck loads flashcards from a JSONL file
func (a *App) LoadDeck(filename string) error {
	a.FlashcardsFile = filename
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
		a.Deck = append(a.Deck, card)
	}
	return nil
}

// SaveNewCard appends the new card to the deck and writes it to the file
func (a *App) SaveNewCard(englishText string) {
	zh, pinyin, err := a.AI.Translate(englishText)
	if err != nil {
		a.Application.Stop()
		fmt.Println("Error translating text:", err)
		return
	}

	newCard := Flashcard{
		ID:      len(a.Deck) + 1,
		English: englishText,
		Chinese: zh,
		Pinyin:  pinyin,
	}
	a.Deck = append(a.Deck, newCard)

	// Append the new card to the flashcards file
	file, err := os.OpenFile(a.FlashcardsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		a.Application.Stop()
		fmt.Println("Error opening flashcards file:", err)
		return
	}
	defer file.Close()

	cardJSON, err := json.Marshal(newCard)
	if err != nil {
		a.Application.Stop()
		fmt.Println("Error marshaling new card:", err)
		return
	}
	if _, err := file.Write(append(cardJSON, '\n')); err != nil {
		a.Application.Stop()
		fmt.Println("Error writing new card to file:", err)
		return
	}

	a.Application.SetRoot(a.MainView, true)
	a.UpdateCardView()
}

// SetupUI initializes the user interface
func (a *App) SetupUI() {
	// Create the main card view
	a.CardView = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetWrap(true)

	// Style the card view
	a.CardView.SetBorder(true).
		SetTitle(" Chinese Learning Cards ").
		SetTitleAlign(tview.AlignCenter)

	// Set up the main view
	a.MainView = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(a.CardView, 0, 2, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	a.UpdateCardView()
}

// UpdateCardView updates the display of the current card
func (a *App) UpdateCardView() {
	if len(a.Deck) == 0 {
		a.CardView.SetText("No cards in deck!")
		return
	}

	card := a.Deck[a.CurrentCardIdx]
	var content strings.Builder
	content.WriteString("\n\n\n") // Add some padding at the top
	content.WriteString(fmt.Sprintf("Card %d/%d (ID: %d)\n\n", a.CurrentCardIdx+1, len(a.Deck), card.ID))

	// Use colors for highlighting
	content.WriteString("[::b]English:[::-]\n")
	content.WriteString("[cyan]" + card.English + "[white]\n\n")

	if a.Revealed {
		content.WriteString("[::b]Chinese:[::-]\n")
		content.WriteString("[yellow]" + card.Chinese + "[white]\n\n")
		content.WriteString("[::b]Pinyin:[::-]\n")
		content.WriteString("[green]" + card.Pinyin + "[white]\n")
	}

	content.WriteString("\n─────────────────────────\n")
	content.WriteString("\nControls:\n")
	content.WriteString("→: Reveal/Next Card  |  n: New Card  |  q: Quit")

	a.CardView.SetText(content.String())
}

// HandleInput processes keyboard input
func (a *App) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if _, ok := a.Application.GetFocus().(*tview.Form); ok {
		return event
	}
	if _, ok := a.Application.GetFocus().(*tview.InputField); ok {
		return event
	}

	switch event.Key() {
	case tcell.KeyRight:
		if !a.Revealed {
			a.Revealed = true
		} else {
			a.Revealed = false
			a.CurrentCardIdx = (a.CurrentCardIdx + 1) % len(a.Deck)
		}
		a.UpdateCardView()
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			a.Application.Stop()
		case 'n':
			a.ShowNewCardDialog()
		}
	}
	return event
}

// ShowNewCardDialog displays the new card input dialog
func (a *App) ShowNewCardDialog() {
	var englishInput *tview.InputField

	form := tview.NewForm()
	englishInput = tview.NewInputField().
		SetLabel("English").
		SetFieldWidth(50)

	form.AddFormItem(englishInput)
	form.AddButton("Save", func() {
		a.SaveNewCard(englishInput.GetText())
	})
	form.AddButton("Cancel", func() {
		a.Application.SetRoot(a.MainView, true)
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

	a.Application.SetRoot(formFlex, true)
}
