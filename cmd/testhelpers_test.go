package cmd

import "strings"

// hasCommand reports whether rootCmd has a registered subcommand whose Use field
// starts with the given name. This allows matching commands with positional-arg
// syntax like "menu <action>" by passing the full Use string or just the verb.
func hasCommand(useOrName string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Use == useOrName {
			return true
		}
		// Match on the first whitespace-delimited token (e.g. "menu <action>" -> "menu")
		if first := firstToken(c.Use); first == useOrName {
			return true
		}
	}
	return false
}

func firstToken(s string) string {
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}
