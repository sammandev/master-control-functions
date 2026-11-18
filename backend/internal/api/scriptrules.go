package api

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Script grammar rules shared with script-audit-mcp (ADR-0012). They mirror
// src/grammar.rs `pattern` and `allowed`; keep both sides in step.

var (
	repeatedTemplate = regexp.MustCompile(`^<[^<>]+>(?:,\s*<[^<>]+>)+,?\s*(?:\.{3}|…)$`)
	characterClass   = regexp.MustCompile(`^\[(?:[A-Za-z0-9](?:-[A-Za-z0-9])?)+\]$`)
)

func isMultiple(defType, cardinality string) bool {
	return cardinality == "MULTIPLE" || defType == "list[int]"
}

// compileValuePattern turns a value_pattern template into the anchored
// expression the MCP validates values with. One template per line; each
// <name> placeholder captures one token.
func compileValuePattern(pattern, defType, cardinality string) (*regexp.Regexp, error) {
	original := strings.TrimSpace(pattern)
	if original == "" {
		return nil, errors.New("value_pattern is missing")
	}
	multi := isMultiple(defType, cardinality)
	var alternatives []string
	for _, line := range strings.Split(original, "\n") {
		expression, err := patternExpression(strings.TrimSpace(line), multi)
		if err != nil {
			return nil, err
		}
		alternatives = append(alternatives, expression)
	}
	return regexp.Compile(`\A(?:` + strings.Join(alternatives, "|") + `)\z`)
}

func patternExpression(original string, multi bool) (string, error) {
	if original == "" {
		return "", errors.New("pattern alternatives must be nonempty")
	}
	source := original
	if repeatedTemplate.MatchString(original) {
		if !multi {
			return "", errors.New("a repeated template requires MULTIPLE cardinality")
		}
		source = "<value>"
	}
	if strings.ContainsAny(source, "[]{}…") || strings.Contains(source, "...") {
		return "", errors.New("optional or undocumented repetition syntax is unsupported")
	}
	var expression strings.Builder
	remaining := source
	count := 0
	for {
		start := strings.IndexByte(remaining, '<')
		if start < 0 {
			break
		}
		expression.WriteString(regexp.QuoteMeta(remaining[:start]))
		tail := remaining[start+1:]
		end := strings.IndexByte(tail, '>')
		if end < 0 {
			return "", errors.New("unterminated pattern placeholder")
		}
		name := tail[:end]
		if name == "" || strings.IndexFunc(name, func(r rune) bool {
			return !(r == '_' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
		}) >= 0 {
			return "", errors.New("placeholder names must contain letters, digits, or underscores")
		}
		expression.WriteString("(.+?)")
		count++
		remaining = tail[end+1:]
	}
	if strings.Contains(remaining, ">") || (multi && count != 1) {
		return "", errors.New("MULTIPLE requires a scalar or uniform repeated placeholder template")
	}
	expression.WriteString(regexp.QuoteMeta(remaining))
	return expression.String(), nil
}

// allowedRule is a parsed allowed_values constraint.
type allowedRule struct {
	kind       string // "", "set", "range", "class"
	set        []string
	lower      float64
	upper      float64
	characters *regexp.Regexp
}

func (r allowedRule) accepts(value string) bool {
	switch r.kind {
	case "set":
		for _, member := range r.set {
			if member == value {
				return true
			}
		}
		return false
	case "range":
		number, err := strconv.ParseFloat(value, 64)
		return err == nil && !math.IsInf(number, 0) && !math.IsNaN(number) &&
			number >= r.lower && number <= r.upper
	case "class":
		return r.characters.MatchString(value)
	default:
		return true
	}
}

// parseAllowed accepts an explicit {a,b} set, a [lower..upper] numeric range,
// or an ASCII character class such as [A-Z0-9]; blank means unconstrained.
func parseAllowed(raw string) (allowedRule, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return allowedRule{}, nil
	}
	if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
		var members []string
		for _, member := range strings.Split(raw[1:len(raw)-1], ",") {
			member = strings.TrimSpace(member)
			if member == "" {
				return allowedRule{}, errors.New("allowed-value set contains an empty member")
			}
			members = append(members, member)
		}
		return allowedRule{kind: "set", set: members}, nil
	}
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		if lowerText, upperText, ok := strings.Cut(raw[1:len(raw)-1], ".."); ok {
			lower, err := strconv.ParseFloat(strings.TrimSpace(lowerText), 64)
			if err != nil {
				return allowedRule{}, errors.New("invalid lower bound")
			}
			upper, err := strconv.ParseFloat(strings.TrimSpace(upperText), 64)
			if err != nil {
				return allowedRule{}, errors.New("invalid upper bound")
			}
			if !math.IsInf(lower, 0) && !math.IsInf(upper, 0) && lower <= upper {
				return allowedRule{kind: "range", lower: lower, upper: upper}, nil
			}
		}
	}
	if characterClass.MatchString(raw) {
		expression, err := regexp.Compile(`\A` + raw + `+\z`)
		if err != nil {
			return allowedRule{}, err
		}
		return allowedRule{kind: "class", characters: expression}, nil
	}
	return allowedRule{}, errors.New("only explicit {sets}, [lower..upper] numeric ranges, and ASCII character classes are supported")
}

// scriptIssue is one finding about a function, field or example, using the
// MCP diagnostic codes so both tools report problems the same way.
type scriptIssue struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// scriptDevice is the device token a script line uses for a Function.
func scriptDevice(source, iface string) string {
	if source == "DLL" {
		return "DLL" + iface
	}
	return iface
}

// checkExampleScript reports structural problems in an example line for any
// of the candidate functions sharing its (source, name). It returns nil when
// at least one candidate accepts the line.
func checkExampleScript(script string, candidates []exampleTarget) []scriptIssue {
	if len(candidates) == 0 {
		return nil
	}
	var best []scriptIssue
	for _, candidate := range candidates {
		issues := checkExampleAgainst(script, candidate)
		if len(issues) == 0 {
			return nil
		}
		if best == nil || len(issues) < len(best) {
			best = issues
		}
	}
	return best
}

// exampleTarget is the part of a Function an example line is checked against.
type exampleTarget struct {
	Source     string
	Interface  string
	Name       string
	Convention *int64
}

func checkExampleAgainst(script string, target exampleTarget) []scriptIssue {
	var issues []scriptIssue
	if strings.ContainsAny(script, "\r\n") {
		return []scriptIssue{{Field: "script", Code: "reserved_character", Message: "an example must be a single script line"}}
	}
	segments := strings.Split(script, "|")
	device := segments[0]
	if base, mode, ok := strings.Cut(device, "#"); ok && target.Source == "DLL" && (mode == "QUICK" || mode == "PROXY") {
		device = base
	}
	if !strings.EqualFold(device, scriptDevice(target.Source, target.Interface)) {
		issues = append(issues, scriptIssue{
			Field:   "script",
			Code:    "device_mismatch",
			Message: "device token '" + segments[0] + "' does not address " + scriptDevice(target.Source, target.Interface),
		})
	}
	if len(segments) < 2 || segments[1] != target.Name {
		issues = append(issues, scriptIssue{
			Field:   "script",
			Code:    "function_mismatch",
			Message: "second segment must be the function name " + target.Name,
		})
	}
	if target.Convention != nil {
		count := len(segments)
		switch *target.Convention {
		case 1:
			if count != 5 {
				issues = append(issues, scriptIssue{Field: "script", Code: "field_count",
					Message: "convention 1 lines have exactly 5 pipe-separated fields, found " + strconv.Itoa(count)})
			}
		case 2:
			// 7 and 9 are historical widths the MCP still accepts.
			if count < 7 || count > 9 {
				issues = append(issues, scriptIssue{Field: "script", Code: "field_count",
					Message: "convention 2 lines have 8 pipe-separated fields, found " + strconv.Itoa(count)})
			}
		}
	}
	return issues
}
