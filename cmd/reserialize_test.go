package cmd

import "testing"

func TestReserializeCmdRegistered(t *testing.T) {
	if !hasCommand("reserialize") {
		t.Fatal("reserialize command not registered")
	}
}

func TestReserializeFlags(t *testing.T) {
	for _, name := range []string{"path", "all", "dry-run"} {
		if reserializeCmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s flag", name)
		}
	}
}
