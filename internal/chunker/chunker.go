package chunker

import (
	"fmt"
	"path/filepath"
	"strings"
)

var extToParser = map[string]func() Parser{
	".go":  func() Parser { return &GoParser{} },
	".ts":  func() Parser { return &TSParser{} },
	".tsx": func() Parser { return &TSParser{} },
	".js":  func() Parser { return &TSParser{} },
	".jsx": func() Parser { return &TSParser{} },
	".py":  func() Parser { return &PythonParser{} },
	".rs":  func() Parser { return &RustParser{} },
}

var extToLang = map[string]string{
	".go": "Go", ".ts": "TypeScript", ".tsx": "TypeScript",
	".js": "JavaScript", ".jsx": "JavaScript",
	".py": "Python", ".rs": "Rust",
}

func Chunk(path string) (*FileChunk, error) {
	ext := strings.ToLower(filepath.Ext(path))

	parserFn, ok := extToParser[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported language for %q", path)
	}

	lang := extToLang[ext]
	parser := parserFn()

	symbols, err := parser.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &FileChunk{
		Path:     path,
		Language: lang,
		Symbols:  symbols,
	}, nil
}
