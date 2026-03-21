package scanner

import (
	"testing"
)

func FuzzParseGitLog(f *testing.F) {
	f.Add("abc1234|feat: init|2026-03-21\ndef5678|fix: bug|2026-03-20\n")
	f.Add("abc|msg with | pipe|2026-01-01\n")
	f.Add("")
	f.Add("invalid line\n")
	f.Add("no-pipe-at-all\n")

	f.Fuzz(func(t *testing.T, output string) {
		parseGitLogOutput(output) // must not panic
	})
}
