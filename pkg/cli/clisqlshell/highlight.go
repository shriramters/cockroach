package clisqlshell

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/scanner"
	"github.com/knz/bubbline/highlight"
)

// sqlHighlighter holds the pre-compiled styles for efficiency.
type sqlHighlighter struct {
	styleKeyword lipgloss.Style
	styleString  lipgloss.Style
	styleNumber  lipgloss.Style
	styleComment lipgloss.Style
	styleIdent   lipgloss.Style
	styleDefault lipgloss.Style
}

// newSQLHighlighter is a constructor that creates the styles only once.
func newSQLHighlighter() *sqlHighlighter {
	yellow := lipgloss.Color("33")
	green := lipgloss.Color("32")
	magenta := lipgloss.Color("35")
	cyan := lipgloss.Color("36")
	gray := lipgloss.Color("244")

	return &sqlHighlighter{
		styleKeyword: lipgloss.NewStyle().Foreground(yellow).Bold(true),
		styleString:  lipgloss.NewStyle().Foreground(green),
		styleNumber:  lipgloss.NewStyle().Foreground(magenta),
		styleComment: lipgloss.NewStyle().Foreground(gray).Italic(true),
		styleIdent:   lipgloss.NewStyle().Foreground(cyan),
		styleDefault: lipgloss.NewStyle(),
	}
}

// In your highlight.go file

// classify returns the correct style for a given token.
func (h *sqlHighlighter) classify(tok scanner.InspectToken) lipgloss.Style {
	// The most reliable way to identify a keyword is to use the lexbase map.
	// We use ToLower() because the keyword map is all lowercase.
	isKeyword := lexbase.GetKeywordID(strings.ToLower(tok.Str)) != lexbase.IDENT

	if isKeyword {
		return h.styleKeyword // Should be Yellow & Bold
	}

	// If it's not a keyword, it's a literal, an identifier, or a symbol.
	switch tok.ID {
	case lexbase.ICONST, lexbase.FCONST:
		return h.styleNumber // Should be Magenta

	case lexbase.SCONST:
		return h.styleString // Should be Green

	case lexbase.IDENT:
		return h.styleIdent // Should be Cyan/Blue

	default:
		// This handles all operators (*, ., ;) and other symbols.
		return h.styleDefault // Should be plain (white)
	}
}

// Highlight method uses the corrected classify logic.
func (h *sqlHighlighter) Highlight(line string) []highlight.Token {
	var tokens []highlight.Token
	var s scanner.SQLScanner
	s.Init(line)

	var lastPos int
	for {
		var lval fakeSym
		s.Scan(&lval)
		if lval.ID() == 0 {
			break
		}

		if int(lval.Pos()) > lastPos {
			tokens = append(tokens, highlight.Token{
				Value: line[lastPos:lval.Pos()],
				Style: h.styleDefault,
			})
		}

		start := int(lval.Pos())
		end := start + len(lval.Str())
		tokStr := line[start:end]

		inspectTok := scanner.InspectToken{
			ID:  lval.ID(),
			Str: tokStr,
		}
		style := h.classify(inspectTok)

		tokens = append(tokens, highlight.Token{
			Value: tokStr,
			Style: style,
		})

		lastPos = end
	}

	if len(line) > lastPos {
		tokens = append(tokens, highlight.Token{
			Value: line[lastPos:],
			Style: h.styleDefault,
		})
	}

	return tokens
}
