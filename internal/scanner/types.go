package scanner

type Report struct {
	Path         string       `json:"path"`
	Tree         *TreeReport  `json:"tree,omitempty"`
	Languages    *LangReport  `json:"languages,omitempty"`
	EntryPoints  []EntryPoint `json:"entry_points,omitempty"`
	Git          *GitReport   `json:"git,omitempty"`
	Dependencies *DepsReport  `json:"dependencies,omitempty"`
	Docs         *DocsReport  `json:"docs,omitempty"`
}

type TreeReport struct {
	TotalFiles int        `json:"total_files"`
	TotalDirs  int        `json:"total_dirs"`
	Entries    []TreeNode `json:"entries"`
}

type TreeNode struct {
	Path     string     `json:"path"`
	IsDir    bool       `json:"is_dir"`
	Files    int        `json:"files,omitempty"`
	Children []TreeNode `json:"children,omitempty"`
}

type LangReport struct {
	TotalFiles int         `json:"total_files"`
	TotalLOC   int         `json:"total_loc"`
	Languages  []LangEntry `json:"languages"`
}

type LangEntry struct {
	Name  string `json:"name"`
	Files int    `json:"files"`
	LOC   int    `json:"loc"`
}

type EntryPoint struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type GitReport struct {
	Branch      string       `json:"branch"`
	DirtyFiles  int          `json:"dirty_files"`
	LastCommits []CommitInfo `json:"last_commits"`
}

type CommitInfo struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
}

type DepsReport struct {
	Manifests []ManifestReport `json:"manifests"`
}

type ManifestReport struct {
	Manager string     `json:"manager"`
	Path    string     `json:"path"`
	Deps    []DepEntry `json:"deps"`
}

type DepEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type DocsReport struct {
	HasReadme   bool     `json:"has_readme"`
	HasClaudeMD bool     `json:"has_claude_md"`
	DocsDir     bool     `json:"docs_dir"`
	Files       []string `json:"files,omitempty"`
}
