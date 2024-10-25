# Chinese Learning Flashcards

A terminal-based flashcard application for learning Chinese, featuring AI-generated translations and pinyin.

## Usage
```bash
go run main.go --api-key=$OPENAI_API_KEY --file=flashcards.jsonl --model=gpt4-mini
```

### Controls
- → (Right Arrow): Reveal card/Next card
- n: Add new card
- q: Quit

## File Format

Uses JSONL format for flashcards:

```json
{"id": 1, "en": "English text", "zh": "Chinese text", "pinyin": "Pinyin text"}
```
