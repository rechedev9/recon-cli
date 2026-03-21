package depgraph

import "testing"

func FuzzParseGoModGraphOutput(f *testing.F) {
	f.Add("example.com/app github.com/foo@v1.0\ngithub.com/foo@v1.0 github.com/bar@v2.0\n")
	f.Add("")
	f.Add("single-field-line\n")
	f.Add("a b\nc d\na c\n") // potential cycle

	f.Fuzz(func(t *testing.T, output string) {
		parseGoModGraphOutput(output, "example.com/app") // must not panic
	})
}
