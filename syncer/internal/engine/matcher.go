package engine

import (
	"path"
	"strings"
)

type matcher struct {
	patterns []compiledPattern
}

type compiledPattern struct {
	pattern  string
	dirOnly  bool
	hasSlash bool
}

func newMatcher(patterns []string) *matcher {
	if len(patterns) == 0 {
		return &matcher{}
	}
	compiled := make([]compiledPattern, 0, len(patterns))
	for _, raw := range patterns {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		dirOnly := strings.HasSuffix(raw, "/") || strings.HasSuffix(raw, "\\")
		raw = strings.TrimSuffix(strings.TrimSuffix(raw, "/"), "\\")
		if raw == "" {
			continue
		}
		normalized := toForwardSlashes(raw)
		compiled = append(compiled, compiledPattern{
			pattern:  normalized,
			dirOnly:  dirOnly,
			hasSlash: strings.Contains(normalized, "/"),
		})
	}
	return &matcher{patterns: compiled}
}

func (m *matcher) ShouldSkip(sectionRelative string, isDir bool) bool {
	if m == nil || len(m.patterns) == 0 {
		return false
	}
	rel := toForwardSlashes(sectionRelative)
	for _, p := range m.patterns {
		candidate := rel
		if !p.hasSlash {
			candidate = path.Base(rel)
		}
		matched, err := path.Match(p.pattern, candidate)
		if err != nil {
			continue
		}
		if matched && (!p.dirOnly || isDir) {
			return true
		}
		if p.dirOnly {
			current := path.Dir(rel)
			for current != "." && current != "/" && current != "" {
				target := current
				if !p.hasSlash {
					target = path.Base(current)
				}
				ok, err := path.Match(p.pattern, target)
				if err == nil && ok {
					return true
				}
				current = path.Dir(current)
			}
		}
	}
	return false
}

func toForwardSlashes(input string) string {
	if input == "" {
		return input
	}
	input = strings.ReplaceAll(input, "\\", "/")
	input = strings.TrimPrefix(input, "./")
	return input
}
