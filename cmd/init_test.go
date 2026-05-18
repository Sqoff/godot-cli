package cmd

import "testing"

func TestInitCmdRegistered(t *testing.T) {
	if !hasCommand("init") {
		t.Fatal("init command not registered")
	}
}

func TestInitFlags(t *testing.T) {
	for _, name := range []string{"force", "enable"} {
		if initCmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s flag", name)
		}
	}
}
