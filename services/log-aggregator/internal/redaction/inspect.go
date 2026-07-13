package redaction

import (
	"encoding/json"
	"sort"
	"strings"
	"unicode"
)

// Inspect decodes one JSON value and returns deterministic policy findings.
// Finding metadata never contains the rejected value.
func Inspect(data []byte, policy Policy) ([]Finding, error) {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return InspectValue(value, policy), nil
}

func InspectValue(value any, policy Policy) []Finding {
	var findings []Finding
	walk(value, "envelope", policy, &findings)
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].NormalizedFieldPath != findings[j].NormalizedFieldPath {
			return findings[i].NormalizedFieldPath < findings[j].NormalizedFieldPath
		}
		if findings[i].RuleID != findings[j].RuleID {
			return findings[i].RuleID < findings[j].RuleID
		}
		return findings[i].ReasonCode < findings[j].ReasonCode
	})
	return findings
}

func walk(value any, path string, policy Policy, findings *[]Finding) {
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			fieldPath := path + "." + normalize(key, policy)
			normalizedKey := normalize(key, policy)
			for _, rule := range policy.Rules {
				if matchesKey(normalizedKey, rule, policy) {
					*findings = append(*findings, Finding{RuleID: rule.ID, NormalizedFieldPath: fieldPath, ReasonCode: rule.ReasonCode})
				}
			}
			walk(child, fieldPath, policy, findings)
		}
	case []any:
		for index, child := range current {
			walk(child, path+"["+itoa(index)+"]", policy, findings)
		}
	case string:
		for _, rule := range policy.Rules {
			if rule.ValuePattern != nil && rule.ValuePattern.MatchString(current) {
				*findings = append(*findings, Finding{RuleID: rule.ID, NormalizedFieldPath: path, ReasonCode: rule.ReasonCode})
			}
		}
	}
}

func matchesKey(key string, rule Rule, policy Policy) bool {
	if len(rule.Keys) == 0 {
		return false
	}
	for _, configured := range rule.Keys {
		if key == normalize(configured, policy) {
			return true
		}
	}
	return false
}

func normalize(value string, policy Policy) string {
	if policy.CaseInsensitive {
		value = strings.ToLower(value)
	}
	if policy.RemoveSeparators {
		var builder strings.Builder
		for _, char := range value {
			remove := false
			for _, separator := range policy.Separators {
				if char == separator || (separator == ' ' && unicode.IsSpace(char)) {
					remove = true
					break
				}
			}
			if !remove {
				builder.WriteRune(char)
			}
		}
		value = builder.String()
	}
	return value
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}
