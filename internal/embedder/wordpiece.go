package embedder

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// WordPiece tokenizes text using the WordPiece algorithm (BERT-style).
type WordPiece struct {
	vocab  map[string]int
	unkID  int
	prefix string // subword prefix, typically "##"
}

// tokenizerJSON matches the HuggingFace tokenizer.json structure (partial).
type tokenizerJSON struct {
	Model struct {
		Vocab    map[string]int `json:"vocab"`
		UnkToken string         `json:"unk_token"`
		Prefix   string         `json:"continuing_subword_prefix"`
	} `json:"model"`
}

// LoadWordPiece loads a WordPiece tokenizer from a HuggingFace tokenizer.json file.
func LoadWordPiece(path string) (*WordPiece, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tokenizer: %w", err)
	}

	var tj tokenizerJSON
	if err := json.Unmarshal(data, &tj); err != nil {
		return nil, fmt.Errorf("parse tokenizer: %w", err)
	}

	unkToken := tj.Model.UnkToken
	if unkToken == "" {
		unkToken = "[UNK]"
	}
	prefix := tj.Model.Prefix
	if prefix == "" {
		prefix = "##"
	}

	unkID, ok := tj.Model.Vocab[unkToken]
	if !ok {
		return nil, fmt.Errorf("unk token %q not in vocab", unkToken)
	}

	return &WordPiece{
		vocab:  tj.Model.Vocab,
		unkID:  unkID,
		prefix: prefix,
	}, nil
}

// Tokenize splits text into WordPiece token IDs.
// Input is lowercased and split on whitespace/punctuation before subword lookup.
func (wp *WordPiece) Tokenize(text string) []int {
	words := splitWords(strings.ToLower(text))
	ids := make([]int, 0, len(words)*2)

	for _, word := range words {
		ids = append(ids, wp.tokenizeWord(word)...)
	}
	return ids
}

// VocabSize returns the number of tokens in the vocabulary.
func (wp *WordPiece) VocabSize() int {
	return len(wp.vocab)
}

// tokenizeWord applies greedy longest-match WordPiece to a single word.
func (wp *WordPiece) tokenizeWord(word string) []int {
	if _, ok := wp.vocab[word]; ok {
		return []int{wp.vocab[word]}
	}

	ids := make([]int, 0, 4)
	start := 0

	for start < len(word) {
		end := len(word)
		matched := false

		for end > start {
			substr := word[start:end]
			if start > 0 {
				substr = wp.prefix + substr
			}

			if id, ok := wp.vocab[substr]; ok {
				ids = append(ids, id)
				start = end
				matched = true
				break
			}
			end--
		}

		if !matched {
			return []int{wp.unkID}
		}
	}
	return ids
}

// splitWords splits text into words on whitespace and punctuation boundaries.
func splitWords(text string) []string {
	var words []string
	var current []rune

	flush := func() {
		if len(current) > 0 {
			words = append(words, string(current))
			current = current[:0]
		}
	}

	for _, r := range text {
		if unicode.IsSpace(r) {
			flush()
		} else if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			flush()
			words = append(words, string(r))
		} else {
			current = append(current, r)
		}
	}
	flush()
	return words
}
