// Package formatter formats NilLang source code into standard canonical style.
package formatter

import (
	"strings"

	"github.com/joysriramsarkar/nilx-framework/compiler/lexer"
)

// Format formats a NilLang source code string with standard 4-space indentation and clean spacing.
func Format(src string) string {
	l := lexer.New("format.nil", src)
	tokens := l.Tokenize()
	if len(tokens) == 0 {
		return src
	}

	lines := strings.Split(src, "\n")
	var result []string
	indent := 0
	inCommentBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, "")
			continue
		}

		if strings.HasPrefix(trimmed, "/*") {
			inCommentBlock = true
		}
		if inCommentBlock {
			result = append(result, strings.Repeat("    ", indent)+trimmed)
			if strings.Contains(trimmed, "*/") {
				inCommentBlock = false
			}
			continue
		}

		if strings.HasPrefix(trimmed, "//") {
			result = append(result, strings.Repeat("    ", indent)+trimmed)
			continue
		}

		// Adjust indent down if line starts with closing brace
		closingCount := countLeadingClosers(trimmed)
		currIndent := indent - closingCount
		if currIndent < 0 {
			currIndent = 0
		}

		formattedLine := formatLine(trimmed)
		result = append(result, strings.Repeat("    ", currIndent)+formattedLine)

		// Compute net change in braces
		opens := strings.Count(trimmed, "{") + strings.Count(trimmed, "(")
		closes := strings.Count(trimmed, "}") + strings.Count(trimmed, ")")
		indent += (opens - closes)
		if indent < 0 {
			indent = 0
		}
	}

	return strings.Join(result, "\n")
}

func countLeadingClosers(s string) int {
	count := 0
	for _, ch := range s {
		if ch == '}' || ch == ')' {
			count++
		} else if ch != ' ' && ch != '\t' {
			break
		}
	}
	return count
}

func formatLine(line string) string {
	// Standardize common operator spacing
	var sb strings.Builder
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if ch == ':' && i+1 < len(runes) && runes[i+1] != ' ' && runes[i+1] != ':' {
			sb.WriteRune(':')
			sb.WriteRune(' ')
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}
