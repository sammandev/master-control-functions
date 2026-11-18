package api

import (
	"database/sql"
	"testing"

	"webprogfndb/internal/store"
)

func TestCompileValuePatternMatchesMCPRules(t *testing.T) {
	cases := []struct {
		pattern, defType, cardinality string
		ok                            bool
		matches, rejects              []string
	}{
		{"<value>", "string", "SINGLE", true, []string{"abc"}, nil},
		{"<ant1>,<ant2>,...", "int", "MULTIPLE", true, []string{"1"}, nil},
		{"<ant1>,<ant2>,…", "list[int]", "SINGLE", true, nil, nil},
		{"<ant1>,<ant2>,...", "int", "SINGLE", false, nil, nil},
		{"", "string", "SINGLE", false, nil, nil},
		{"<a>\n\n<b>", "string", "SINGLE", false, nil, nil},
		{"<lo>~<hi>", "string", "SINGLE", true, []string{"1~2"}, []string{"12"}},
		{"<lo>~<hi>\r\n<only>", "string", "SINGLE", true, []string{"1~2", "x"}, nil},
		{"[<value>]", "string", "SINGLE", false, nil, nil},
		{"<value", "string", "SINGLE", false, nil, nil},
		{"value>", "string", "SINGLE", false, nil, nil},
		{"<bad-name>", "string", "SINGLE", false, nil, nil},
		{"<a>:<b>", "int", "MULTIPLE", false, nil, nil},
	}
	for _, tc := range cases {
		expression, err := compileValuePattern(tc.pattern, tc.defType, tc.cardinality)
		if (err == nil) != tc.ok {
			t.Fatalf("%q (%s/%s): err = %v, want ok=%v", tc.pattern, tc.defType, tc.cardinality, err, tc.ok)
		}
		for _, v := range tc.matches {
			if !expression.MatchString(v) {
				t.Fatalf("%q should match %q", tc.pattern, v)
			}
		}
		for _, v := range tc.rejects {
			if expression.MatchString(v) {
				t.Fatalf("%q should reject %q", tc.pattern, v)
			}
		}
	}
}

func TestParseAllowedMatchesMCPRules(t *testing.T) {
	cases := []struct {
		raw      string
		ok       bool
		accepts  []string
		rejected []string
	}{
		{"", true, []string{"anything"}, nil},
		{"{1,2,4}", true, []string{"2"}, []string{"3"}},
		{"{1,,2}", false, nil, nil},
		{"{}", false, nil, nil},
		{"[-10..30.5]", true, []string{"-10", "30.5", "0"}, []string{"31", "x"}},
		{"[5..1]", false, nil, nil},
		{"[A-Z0-9]", true, []string{"AB12"}, []string{"ab"}},
		{"1,2,3", false, nil, nil},
	}
	for _, tc := range cases {
		rule, err := parseAllowed(tc.raw)
		if (err == nil) != tc.ok {
			t.Fatalf("%q: err = %v, want ok=%v", tc.raw, err, tc.ok)
		}
		for _, v := range tc.accepts {
			if !rule.accepts(v) {
				t.Fatalf("%q should accept %q", tc.raw, v)
			}
		}
		for _, v := range tc.rejected {
			if rule.accepts(v) {
				t.Fatalf("%q should reject %q", tc.raw, v)
			}
		}
	}
}

func TestCheckExampleScript(t *testing.T) {
	two := int64(2)
	one := int64(1)
	tx := exampleTarget{"DLL", "WIFI_TEST", "TX_MULTI_VERIFICATION", &two}
	ssh := exampleTarget{"BUILT-IN", "SSH", "OPEN", &one}

	if issues := checkExampleScript("DLLWIFI_TEST#QUICK|TX_MULTI_VERIFICATION|ANT=1|POWER=[1..2]|10|NEXT|FAIL|E1", []exampleTarget{tx}); issues != nil {
		t.Fatalf("valid convention-2 line flagged: %v", issues)
	}
	if issues := checkExampleScript("SSH|OPEN|host|NULL|5", []exampleTarget{ssh}); issues != nil {
		t.Fatalf("valid convention-1 line flagged: %v", issues)
	}
	issues := checkExampleScript("DLLIOT|OTHER|x|y|1", []exampleTarget{tx})
	codes := map[string]bool{}
	for _, issue := range issues {
		codes[issue.Code] = true
	}
	for _, want := range []string{"device_mismatch", "function_mismatch", "field_count"} {
		if !codes[want] {
			t.Fatalf("missing %s in %v", want, issues)
		}
	}
	if issues := checkExampleScript("SSH#QUICK|OPEN|host|NULL|5", []exampleTarget{ssh}); issues == nil {
		t.Fatalf("#QUICK on a BUILT-IN device must be flagged")
	}
	// Any candidate accepting the line clears it.
	iot := exampleTarget{"DLL", "IOT", "TX_MULTI_VERIFICATION", &two}
	if issues := checkExampleScript("DLLIOT|TX_MULTI_VERIFICATION|A=1|NULL|1|N|F|E", []exampleTarget{tx, iot}); issues != nil {
		t.Fatalf("second candidate should accept: %v", issues)
	}
}

func TestSupportIssuesFollowGrammarAndAudits(t *testing.T) {
	f := store.Function{Source: "DLL", Interface: "WIFI_TEST", Name: "F", Convention: sql.NullInt64{Int64: 2, Valid: true}}
	def := &catalogDefinition{Name: "ANT", Type: "int", Cardinality: "SINGLE", ValuePattern: "<ant>"}
	field := readinessField{Name: "ANT", Definition: true, definition: def}

	if issues := supportIssues(f, nil, nil, nil, true); len(issues) != 1 || issues[0].Code != "unsupported_grammar" {
		t.Fatalf("missing grammar: %v", issues)
	}
	bundleOnly := &store.FunctionGrammar{Encoding: "named", Scope: "bundle"}
	field.ScriptKey = "bundle"
	if issues := supportIssues(f, bundleOnly, []readinessField{field}, nil, false); len(issues) != 1 {
		t.Fatalf("bundle grammar must not compose lines: %v", issues)
	}
	if issues := supportIssues(f, bundleOnly, []readinessField{field}, nil, true); len(issues) != 0 {
		t.Fatalf("bundle should be ready: %v", issues)
	}
	line := &store.FunctionGrammar{Encoding: "named", Scope: "line"}
	if issues := supportIssues(f, line, []readinessField{field}, nil, false); len(issues) != 1 {
		t.Fatalf("bundle-only key must block line composition: %v", issues)
	}
	field.ScriptKey = ""
	if issues := supportIssues(f, line, []readinessField{field}, nil, true); len(issues) != 1 {
		t.Fatalf("unaudited key must be flagged: %v", issues)
	}
	raw := &store.FunctionGrammar{Encoding: "raw", Scope: "line"}
	if issues := supportIssues(f, raw, []readinessField{field}, nil, false); len(issues) != 1 {
		t.Fatalf("raw grammar with convention 2 must be flagged: %v", issues)
	}
}
