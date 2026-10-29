package migrator

import (
	"errors"
	"strings"
)

// SplitStatements splits a SQL script into individual statements.
//
// Naively splitting on ";" breaks as soon as a semicolon appears inside a
// string literal, a quoted identifier, a comment or a dollar-quoted block
// (function bodies). This splitter tracks those contexts instead.
func SplitStatements(script string) ([]string, error) {
	var (
		statements []string
		current    strings.Builder
	)

	runes := []rune(script)

	for i := 0; i < len(runes); i++ {
		char := runes[i]

		switch {
		case char == '-' && peek(runes, i+1) == '-':
			// Line comment: copy it verbatim, it does not affect splitting.
			end := indexOrEnd(runes, i, "\n")
			current.WriteString(string(runes[i:end]))
			i = end - 1

		case char == '/' && peek(runes, i+1) == '*':
			end, err := skipBlockComment(runes, i)
			if err != nil {
				return nil, err
			}

			current.WriteString(string(runes[i:end]))
			i = end - 1

		case char == '\'' || char == '"':
			end, err := skipQuoted(runes, i, char)
			if err != nil {
				return nil, err
			}

			current.WriteString(string(runes[i:end]))
			i = end - 1

		case char == '$':
			end, ok, err := skipDollarQuoted(runes, i)
			if err != nil {
				return nil, err
			}

			if !ok {
				current.WriteRune(char)

				continue
			}

			current.WriteString(string(runes[i:end]))
			i = end - 1

		case char == ';':
			if statement := strings.TrimSpace(current.String()); statement != "" {
				statements = append(statements, statement)
			}

			current.Reset()

		default:
			current.WriteRune(char)
		}
	}

	if statement := strings.TrimSpace(current.String()); statement != "" {
		statements = append(statements, statement)
	}

	return statements, nil
}

func peek(runes []rune, index int) rune {
	if index < 0 || index >= len(runes) {
		return 0
	}

	return runes[index]
}

func indexOrEnd(runes []rune, from int, needle string) int {
	index := strings.Index(string(runes[from:]), needle)
	if index < 0 {
		return len(runes)
	}

	return from + index
}

// skipBlockComment returns the index just past the closing "*/".
func skipBlockComment(runes []rune, start int) (int, error) {
	for i := start + 2; i < len(runes); i++ {
		if runes[i] == '*' && peek(runes, i+1) == '/' {
			return i + 2, nil
		}
	}

	return 0, errors.New("unterminated block comment")
}

// skipQuoted returns the index just past the closing quote. Doubled quotes are
// treated as escaped characters.
func skipQuoted(runes []rune, start int, quote rune) (int, error) {
	for i := start + 1; i < len(runes); i++ {
		if runes[i] != quote {
			continue
		}

		if peek(runes, i+1) == quote {
			i++

			continue
		}

		return i + 1, nil
	}

	return 0, errors.New("unterminated quoted literal")
}

// skipDollarQuoted handles PostgreSQL dollar quoting: $tag$ ... $tag$.
// ok is false when the dollar sign does not start a dollar-quoted block.
func skipDollarQuoted(runes []rune, start int) (end int, ok bool, err error) {
	tagEnd := start + 1

	for tagEnd < len(runes) && (isIdentifierRune(runes[tagEnd])) {
		tagEnd++
	}

	if tagEnd >= len(runes) || runes[tagEnd] != '$' {
		return 0, false, nil
	}

	tag := string(runes[start : tagEnd+1])
	rest := string(runes[tagEnd+1:])

	offset := strings.Index(rest, tag)
	if offset < 0 {
		return 0, false, errors.New("unterminated dollar-quoted string")
	}

	return tagEnd + 1 + offset + len(tag), true, nil
}

func isIdentifierRune(char rune) bool {
	return char == '_' ||
		(char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}
