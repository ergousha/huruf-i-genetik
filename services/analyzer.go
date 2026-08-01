package services

import (
	"strings"
	"unicode"
)

// nonConnectingChars are letters that do not connect to the letter after them.
// Alif, Dal, Dhal, Ra, Zay, Waw
var nonConnectingChars = map[string]bool{
	"ا": true,
	"د": true,
	"ذ": true,
	"ر": true,
	"ز": true,
	"و": true,
	"ء": true, // Hemze doesn't usually connect like standard letters
}

// medChars are letters that can act as elongation (Madd).
var medChars = map[string]bool{
	"ا": true,
	"و": true,
	"ي": true,
}

// AnalyzerService processes text to find relations.
type AnalyzerService struct{}

func NewAnalyzerService() *AnalyzerService {
	return &AnalyzerService{}
}

// CleanText removes non-arabic characters and diacritics (tashkeel).
func (s *AnalyzerService) CleanText(input string) []string {
	var chars []string
	for _, runeValue := range input {
		// Basic Arabic block check, ignoring combining marks (tashkeel/harakat)
		if unicode.Is(unicode.Arabic, runeValue) && !unicode.Is(unicode.Mn, runeValue) {
			chars = append(chars, string(runeValue))
		}
	}
	// For some environments, spaces might be needed, but we focus on consecutive letters.
	return chars
}

// ParseRelations analyzes sequential characters and determines their connection type.
func (s *AnalyzerService) ParseRelations(chars []string) []Relation {
	var relations []Relation

	for i := 0; i < len(chars)-1; i++ {
		curr := chars[i]
		next := chars[i+1]

		// Skip spaces if they sneaked in, though CleanText shouldn't include them
		if strings.TrimSpace(curr) == "" || strings.TrimSpace(next) == "" {
			continue
		}

		relType := "VASIL" // Default is connect

		if nonConnectingChars[curr] {
			relType = "FASIL"
		}

		// Simplified Med rule: If next is a Med letter, treat it as MED relation.
		// (A true Med rule would check preceding vowel diacritics, but we are working with unvoweled structure for now).
		if medChars[next] {
			relType = "MED"
		}

		relations = append(relations, Relation{
			From: curr,
			To:   next,
			Type: relType,
		})
	}
	return relations
}

type Relation struct {
	From string
	To   string
	Type string // VASIL, FASIL, MED
}
