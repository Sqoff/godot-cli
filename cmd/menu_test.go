package cmd

import "testing"

func TestMenuCmdRegistered(t *testing.T) {
	if !hasCommand("menu <action>") {
		t.Fatal("menu command not registered")
	}
}
