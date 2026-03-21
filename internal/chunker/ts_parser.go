package chunker

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type TSParser struct{}

var tsPatterns = []struct {
	re   *regexp.Regexp
	kind string
}{
	{regexp.MustCompile(`^export\s+default\s+(class|function|abstract\s+class)\s+(\w+)`), ""},
	{regexp.MustCompile(`^export\s+(function|class|abstract\s+class|interface|type|enum|const|let|var)\s+(\w+)`), ""},
	{regexp.MustCompile(`^(function|class|abstract\s+class|interface|type|enum)\s+(\w+)`), ""},
	{regexp.MustCompile(`^(const|let|var)\s+(\w+)\s*[=:]`), ""},
}

func (p *TSParser) Parse(path string) ([]Symbol, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open ts file: %w", err)
	}
	defer f.Close()

	var symbols []Symbol
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())

		for _, pat := range tsPatterns {
			matches := pat.re.FindStringSubmatch(text)
			if matches == nil {
				continue
			}

			kindStr := matches[1]
			name := matches[2]

			kind := normalizeKind(kindStr)

			symbols = append(symbols, Symbol{
				Name:      name,
				Kind:      kind,
				Line:      line,
				Signature: text,
			})
			break
		}
	}

	return symbols, sc.Err()
}

func normalizeKind(s string) string {
	s = strings.TrimSpace(s)
	switch {
	case strings.Contains(s, "class"):
		return "class"
	case s == "function":
		return "function"
	case s == "interface":
		return "interface"
	case s == "type":
		return "type"
	case s == "enum":
		return "type"
	case s == "const" || s == "let" || s == "var":
		return "const"
	default:
		return s
	}
}
