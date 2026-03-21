package chunker

type Symbol struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`      // "function", "type", "struct", "interface", "class", "export", "const", "var"
	Line      int    `json:"line"`
	Signature string `json:"signature"`
}

type FileChunk struct {
	Path     string   `json:"path"`
	Language string   `json:"language"`
	Symbols  []Symbol `json:"symbols"`
}

type Parser interface {
	Parse(path string) ([]Symbol, error)
}
