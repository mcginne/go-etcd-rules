package rules

import (
	"strings"
)

type ruleManager struct {
	constraints       map[string]constraint
	currentIndex      int
	rulesBySlashCount map[int]map[DynamicRule]int
	prefixes          map[string]string
}

func newRuleManager(constraints map[string]constraint) ruleManager {
	rm := ruleManager{
		rulesBySlashCount: map[int]map[DynamicRule]int{},
		prefixes:          map[string]string{},
		constraints:       constraints,
		currentIndex:      0,
	}
	return rm
}

func (rm *ruleManager) getStaticRules(key string, value *string) map[staticRule]int {
	slashCount := strings.Count(key, "/")
	out := make(map[staticRule]int)
	rules, ok := rm.rulesBySlashCount[slashCount]
	if ok {
		for rule, index := range rules {
			sRule, _, inScope := rule.makeStaticRule(key, value)
			if inScope && sRule.satisfiable(key, value) {
				out[sRule] = index
			}
		}
	}
	return out
}

func (rm *ruleManager) addRule(rule DynamicRule) int {
	for _, pattern := range rule.getPatterns() {
		slashCount := strings.Count(pattern, "/")
		rules, ok := rm.rulesBySlashCount[slashCount]
		if !ok {
			rules = map[DynamicRule]int{}
			rm.rulesBySlashCount[slashCount] = rules
		}
		rules[rule] = rm.currentIndex
	}
	for _, prefix := range rule.getPrefixesWithConstraints(rm.constraints) {
		rm.prefixes[prefix] = ""
	}
	rm.prefixes = reducePrefixes(rm.prefixes)
	lastIndex := rm.currentIndex
	rm.currentIndex = rm.currentIndex + 1
	return lastIndex
}

// Removes any path prefixes that have other path prefixes as
// string prefixes
func reducePrefixes(prefixes map[string]string) map[string]string {
	out := map[string]string{}
	sorted := sortPrefixesByLength(prefixes)
	for _, prefix := range sorted {
		add := true
		for addedPrefix := range out {
			if strings.HasPrefix(prefix, addedPrefix) {
				add = false
			}
		}
		if add {
			out[prefix] = ""
		}
	}
	return out
}

// Sorts prefixes shortest to longest
func sortPrefixesByLength(prefixes map[string]string) []string {
	out := []string{}
	for prefix := range prefixes {
		out = append(out, prefix)
	}
	for i := 1; i < len(out); i++ {
		x := out[i]
		j := i - 1
		for j >= 0 && len(out[j]) > len(x) {
			out[j+1] = out[j]
			j = j - 1
		}
		out[j+1] = x
	}
	return out
}
