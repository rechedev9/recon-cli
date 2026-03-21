package scanner

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var extToLang = map[string]string{
	".go":   "Go",
	".ts":   "TypeScript",
	".tsx":  "TypeScript",
	".js":   "JavaScript",
	".jsx":  "JavaScript",
	".py":   "Python",
	".rs":   "Rust",
	".md":   "Markdown",
	".yaml": "YAML",
	".yml":  "YAML",
	".json": "JSON",
	".sh":   "Shell",
	".sql":  "SQL",
	".html": "HTML",
	".css":  "CSS",
	".toml": "TOML",
}

type langScanner struct{}

func (s *langScanner) Scan(ctx context.Context, root string) (*LangReport, error) {
	counts := make(map[string]*LangEntry)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() && shouldSkip(d.Name()) && path != root {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		lang, ok := extToLang[ext]
		if !ok {
			return nil
		}

		loc, err := countLOC(path)
		if err != nil {
			return nil
		}

		entry, exists := counts[lang]
		if !exists {
			entry = &LangEntry{Name: lang}
			counts[lang] = entry
		}
		entry.Files++
		entry.LOC += loc
		return nil
	})
	if err != nil {
		return nil, err
	}

	report := &LangReport{}
	for _, entry := range counts {
		report.Languages = append(report.Languages, *entry)
		report.TotalFiles += entry.Files
		report.TotalLOC += entry.LOC
	}

	sort.Slice(report.Languages, func(i, j int) bool {
		return report.Languages[i].LOC > report.Languages[j].LOC
	})

	return report, nil
}

func countLOC(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) != "" {
			count++
		}
	}
	return count, sc.Err()
}
