package cmd

import "testing"

func TestNodeCmdRegistered(t *testing.T) {
	if !hasCommand("node") {
		t.Fatal("node command not registered")
	}
}

func TestNodeSubcommands(t *testing.T) {
	subs := map[string]bool{}
	for _, c := range nodeCmd.Commands() {
		subs[firstToken(c.Use)] = true
	}
	for _, want := range []string{"tree", "get", "set"} {
		if !subs[want] {
			t.Errorf("missing subcommand: %s", want)
		}
	}
}

func TestNodeTreeFlags(t *testing.T) {
	for _, name := range []string{"depth", "props"} {
		if nodeTreeCmd.Flags().Lookup(name) == nil {
			t.Errorf("missing tree --%s flag", name)
		}
	}
}
