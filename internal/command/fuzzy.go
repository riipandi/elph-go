package command

import "strings"

// fuzzyScore returns a relevance score for query against target.
// Higher is better; -1 means no subsequence match.
func fuzzyScore(query, target string) int {
	query = strings.ToLower(strings.TrimSpace(query))
	target = strings.ToLower(target)
	if query == "" {
		return 0
	}
	if target == "" {
		return -1
	}

	qi := 0
	score := 0
	prev := -2
	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if target[ti] == query[qi] {
			score++
			if ti == 0 {
				score += 8
			}
			if ti == prev+1 {
				score += 4
			}
			prev = ti
			qi++
		}
	}
	if qi != len(query) {
		return -1
	}
	return score
}

func commandScore(query string, cmd SlashCommand) int {
	best := -1
	for _, term := range commandTerms(cmd) {
		if score := fuzzyScore(query, term); score > best {
			best = score
		}
	}
	return best
}

func commandTerms(cmd SlashCommand) []string {
	terms := make([]string, 0, 2+len(cmd.Aliases))
	terms = append(terms, cmd.Name)
	terms = append(terms, cmd.Aliases...)
	if cmd.Description != "" {
		terms = append(terms, cmd.Description)
	}
	return terms
}
