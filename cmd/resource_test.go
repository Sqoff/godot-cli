package cmd

import "testing"

func TestResourceCmdRegistered(t *testing.T) {
	if !hasCommand("resource") {
		t.Fatal("resource command not registered")
	}
}

func TestResourceSubcommands(t *testing.T) {
	subs := map[string]bool{}
	for _, c := range resourceCmd.Commands() {
		subs[firstToken(c.Use)] = true
	}
	for _, want := range []string{"find", "info"} {
		if !subs[want] {
			t.Errorf("missing subcommand: %s", want)
		}
	}
}
