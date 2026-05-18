package cmd

import "testing"

func TestScreenshotCmdRegistered(t *testing.T) {
	if !hasCommand("screenshot") {
		t.Fatal("screenshot command not registered")
	}
}

func TestScreenshotFlags(t *testing.T) {
	for _, name := range []string{"output", "target"} {
		if screenshotCmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s flag", name)
		}
	}
}
