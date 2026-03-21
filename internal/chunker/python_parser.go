package chunker

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type PythonParser struct{}

var pyPatterns = []struct {
	re   *regexp.Regexp
	kind string
}{
	{regexp.MustCompile(`^class\s+(\w+)`), "class"},
	{regexp.MustCompile(`^(?:async\s+)?def\s+(\w+)`), "function"},
	{regexp.MustCompile(`^([A-Z][A-Z_0-9]*)\s*=`), "const"},
}

func (p *PythonParser) Parse(path string) ([]Symbol, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open python file: %w", err)
	}
	defer f.Close()

	var symbols []Symbol
	sc := bufio.NewScanner(f)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		text := sc.Text()

		// Only match top-level definitions (no leading whitespace)
		if len(text) > 0 && (text[0] == ' ' || text[0] == '\t') {
			continue
		}

		trimmed := strings.TrimSpace(text)

		for _, pat := range pyPatterns {
			matches := pat.re.FindStringSubmatch(trimmed)
			if matches == nil {
				continue
			}

			symbols = append(symbols, Symbol{
				Name:      matches[1],
				Kind:      pat.kind,
				Line:      lineNum,
				Signature: trimmed,
			})
			break
		}
	}

	return symbols, sc.Err()
}
