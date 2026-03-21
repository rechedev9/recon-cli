package chunker

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

type RustParser struct{}

var rustPatterns = []struct {
	re   *regexp.Regexp
	kind string
}{
	{regexp.MustCompile(`^(?:pub\s+)?fn\s+(\w+)`), "function"},
	{regexp.MustCompile(`^(?:pub\s+)?struct\s+(\w+)`), "struct"},
	{regexp.MustCompile(`^(?:pub\s+)?enum\s+(\w+)`), "type"},
	{regexp.MustCompile(`^(?:pub\s+)?trait\s+(\w+)`), "interface"},
	{regexp.MustCompile(`^(?:pub\s+)?type\s+(\w+)`), "type"},
	{regexp.MustCompile(`^(?:pub\s+)?const\s+(\w+)`), "const"},
	{regexp.MustCompile(`^(?:pub\s+)?static\s+(\w+)`), "const"},
	{regexp.MustCompile(`^(?:pub\s+)?mod\s+(\w+)`), "module"},
	{regexp.MustCompile(`^impl\s+(\w+)`), "impl"},
}

func (p *RustParser) Parse(path string) ([]Symbol, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open rust file: %w", err)
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

		for _, pat := range rustPatterns {
			matches := pat.re.FindStringSubmatch(text)
			if matches == nil {
				continue
			}

			symbols = append(symbols, Symbol{
				Name:      matches[1],
				Kind:      pat.kind,
				Line:      lineNum,
				Signature: text,
			})
			break
		}
	}

	return symbols, sc.Err()
}
