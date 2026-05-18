package cmd

import "testing"

func TestPauseCmdRegistered(t *testing.T) {
	if !hasCommand("pause") {
		t.Fatal("pause command not registered")
	}
}

func TestPauseFlags(t *testing.T) {
	if pauseCmd.Flags().Lookup("on") == nil {
		t.Error("missing --on flag")
	}
	if pauseCmd.Flags().Lookup("off") == nil {
		t.Error("missing --off flag")
	}
}
