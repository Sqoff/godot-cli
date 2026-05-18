package cmd

import "testing"

func TestScriptCmdRegistered(t *testing.T) {
	if !hasCommand("script") {
		t.Fatal("script command not registered")
	}
}

func TestScriptFlags(t *testing.T) {
	if scriptCmd.Flags().Lookup("method") == nil {
		t.Error("missing --method flag")
	}
}
