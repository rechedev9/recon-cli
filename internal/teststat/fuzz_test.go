package teststat

import "testing"

func FuzzParseGoTestJSON(f *testing.F) {
	f.Add(`{"Action":"run","Package":"app","Test":"TestFoo"}
{"Action":"pass","Package":"app","Test":"TestFoo","Elapsed":0.1}
{"Action":"pass","Package":"app","Elapsed":1.0}
`)
	f.Add(`{"Action":"fail","Package":"app","Test":"TestBar","Elapsed":0.1}`)
	f.Add("")
	f.Add("not json\n")
	f.Add(`{"Action":"unknown"}`)

	f.Fuzz(func(t *testing.T, output string) {
		parseGoTestJSON(output) // must not panic
	})
}

func FuzzParseJestJSON(f *testing.F) {
	f.Add(`{"numPassedTests":1,"numFailedTests":0,"numPendingTests":0,"numTotalTests":1,"testResults":[]}`)
	f.Add(`{}`)
	f.Add("")
	f.Add("not json")

	f.Fuzz(func(t *testing.T, output string) {
		parseJestJSON(output) // must not panic
	})
}

func FuzzParsePytestOutput(f *testing.F) {
	f.Add("3 passed, 1 failed in 2.5s\n")
	f.Add("FAILED tests/test_api.py::test_foo - AssertionError\n5 passed in 1.0s\n")
	f.Add("")
	f.Add("no summary line")

	f.Fuzz(func(t *testing.T, output string) {
		parsePytestOutput(output) // must not panic
	})
}
