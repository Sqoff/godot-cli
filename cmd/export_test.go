package cmd

import "testing"

func TestExportCmdRegistered(t *testing.T) {
	if !hasCommand("export") {
		t.Fatal("export command not registered")
	}
}

func TestExportFlags(t *testing.T) {
	for _, name := range []string{"preset", "output", "debug", "godot"} {
		if exportCmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s flag", name)
		}
	}
}
