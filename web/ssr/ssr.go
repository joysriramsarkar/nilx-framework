package ssr

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

type Component interface {
	RenderHTML() string
}

type SSRRenderer struct {
	Title string
	Theme string
}

func New(title string) *SSRRenderer {
	return &SSRRenderer{Title: title, Theme: "dark"}
}

func (s *SSRRenderer) RenderPage(content string, state map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	sb.WriteString("  <meta charset=\"UTF-8\">\n")
	sb.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", html.EscapeString(s.Title)))
	sb.WriteString("  <style>\n")
	sb.WriteString("    body { font-family: system-ui, sans-serif; background: #0a0a0f; color: #f8fafc; margin: 0; padding: 24px; }\n")
	sb.WriteString("    .container { max-width: 900px; margin: 0 auto; }\n")
	sb.WriteString("  </style>\n</head>\n<body>\n")
	sb.WriteString("  <div id=\"alap-root\" class=\"container\">\n")
	sb.WriteString(content)
	sb.WriteString("\n  </div>\n")

	if state != nil {
		stateJSON, _ := json.Marshal(state)
		sb.WriteString(fmt.Sprintf("\n  <script id=\"__NILANG_STATE__\" type=\"application/json\">%s</script>\n", string(stateJSON)))
		sb.WriteString("  <script>window.__NILANG_INITIAL_STATE__ = JSON.parse(document.getElementById('__NILANG_STATE__').textContent);</script>\n")
	}

	sb.WriteString("</body>\n</html>")
	return sb.String()
}
