package cmd

import "testing"

func TestProfilerCmdRegistered(t *testing.T) {
	if !hasCommand("profiler") {
		t.Fatal("profiler command not registered")
	}
}

func TestProfilerSubcommands(t *testing.T) {
	subs := map[string]bool{}
	for _, c := range profilerCmd.Commands() {
		subs[firstToken(c.Use)] = true
	}
	for _, want := range []string{"status", "enable", "disable", "clear", "hierarchy"} {
		if !subs[want] {
			t.Errorf("missing subcommand: %s", want)
		}
	}
}
