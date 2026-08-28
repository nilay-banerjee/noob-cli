package installer

import "testing"

func TestBrewListName(t *testing.T) {
	cases := map[string]string{
		"git":                       "git",
		"claude-code@latest":        "claude-code",
		"nikitabobko/tap/aerospace": "aerospace",
		"jnsahaj/lumen/lumen":       "lumen",
		"postgresql@18":             "postgresql",
	}
	for in, want := range cases {
		if got := brewListName(in); got != want {
			t.Errorf("brewListName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestElevate(t *testing.T) {
	got := elevate("apt-get", "update")
	if runningAsRoot() {
		if len(got) != 2 || got[0] != "apt-get" {
			t.Errorf("as root, elevate should not prepend sudo: %v", got)
		}
	} else {
		if len(got) != 3 || got[0] != "sudo" {
			t.Errorf("as non-root, elevate should prepend sudo: %v", got)
		}
	}
}
