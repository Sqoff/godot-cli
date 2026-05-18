package cmd

import "testing"

func TestSettingCmdRegistered(t *testing.T) {
	if !hasCommand("setting") {
		t.Fatal("setting command not registered")
	}
}

func TestSettingSubcommands(t *testing.T) {
	subs := map[string]bool{}
	for _, c := range settingCmd.Commands() {
		subs[firstToken(c.Use)] = true
	}
	for _, want := range []string{"get", "set", "list"} {
		if !subs[want] {
			t.Errorf("missing subcommand: %s", want)
		}
	}
}

func TestSettingListFlags(t *testing.T) {
	for _, name := range []string{"prefix", "all"} {
		if settingListCmd.Flags().Lookup(name) == nil {
			t.Errorf("missing list --%s flag", name)
		}
	}
}
