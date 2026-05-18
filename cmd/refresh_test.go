package cmd

import "testing"

func TestRefreshCmdRegistered(t *testing.T) {
	if !hasCommand("refresh") {
		t.Fatal("refresh command not registered")
	}
}

func TestRefreshFlags(t *testing.T) {
	if refreshCmd.Flags().Lookup("sources") == nil {
		t.Error("missing --sources flag")
	}
}
