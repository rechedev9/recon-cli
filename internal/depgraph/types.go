package depgraph

type DepNode struct {
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Children []DepNode `json:"children,omitempty"`
}

type GraphReport struct {
	Manager string    `json:"manager"`
	Root    string    `json:"root"`
	Nodes   []DepNode `json:"nodes"`
}

type GraphBuilder interface {
	Build(root string) (*GraphReport, error)
}
