package chunker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type GoParser struct{}

func (p *GoParser) Parse(path string) ([]Symbol, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse go file: %w", err)
	}

	var symbols []Symbol

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			name := d.Name.Name
			if d.Recv != nil && len(d.Recv.List) > 0 {
				recvType := exprString(d.Recv.List[0].Type)
				name = strings.TrimPrefix(recvType, "*") + "." + name
			}
			sig := funcSignature(d)
			symbols = append(symbols, Symbol{
				Name:      name,
				Kind:      "function",
				Line:      fset.Position(d.Pos()).Line,
				Signature: sig,
			})

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					kind := "type"
					switch s.Type.(type) {
					case *ast.StructType:
						kind = "struct"
					case *ast.InterfaceType:
						kind = "interface"
					}
					symbols = append(symbols, Symbol{
						Name:      s.Name.Name,
						Kind:      kind,
						Line:      fset.Position(s.Pos()).Line,
						Signature: fmt.Sprintf("type %s %s", s.Name.Name, kind),
					})

				case *ast.ValueSpec:
					kind := "var"
					if d.Tok == token.CONST {
						kind = "const"
					}
					for _, name := range s.Names {
						symbols = append(symbols, Symbol{
							Name:      name.Name,
							Kind:      kind,
							Line:      fset.Position(name.Pos()).Line,
							Signature: fmt.Sprintf("%s %s", kind, name.Name),
						})
					}
				}
			}
		}
	}

	return symbols, nil
}

func funcSignature(d *ast.FuncDecl) string {
	var b strings.Builder
	b.WriteString("func ")
	if d.Recv != nil && len(d.Recv.List) > 0 {
		b.WriteString("(")
		b.WriteString(exprString(d.Recv.List[0].Type))
		b.WriteString(") ")
	}
	b.WriteString(d.Name.Name)
	b.WriteString("(")
	if d.Type.Params != nil {
		params := make([]string, 0)
		for _, p := range d.Type.Params.List {
			typeStr := exprString(p.Type)
			if len(p.Names) == 0 {
				params = append(params, typeStr)
			}
			for _, name := range p.Names {
				params = append(params, name.Name+" "+typeStr)
			}
		}
		b.WriteString(strings.Join(params, ", "))
	}
	b.WriteString(")")
	if d.Type.Results != nil && len(d.Type.Results.List) > 0 {
		results := make([]string, 0)
		for _, r := range d.Type.Results.List {
			results = append(results, exprString(r.Type))
		}
		if len(results) == 1 {
			b.WriteString(" " + results[0])
		} else {
			b.WriteString(" (" + strings.Join(results, ", ") + ")")
		}
	}
	return b.String()
}

func exprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprString(e.X)
	case *ast.SelectorExpr:
		return exprString(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprString(e.Elt)
	case *ast.MapType:
		return "map[" + exprString(e.Key) + "]" + exprString(e.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + exprString(e.Elt)
	default:
		return "any"
	}
}
