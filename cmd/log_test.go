package cmd

import "testing"

func TestLogCmdRegistered(t *testing.T) {
	if !hasCommand("log") {
		t.Fatal("log command not registered")
	}
}

func TestLogFlags(t *testing.T) {
	for _, name := range []string{"lines", "type", "filter", "clear", "follow"} {
		if logCmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s flag", name)
		}
	}
}
