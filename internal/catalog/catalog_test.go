package catalog

import (
	"strings"
	"testing"
)

func names(items []Item) map[string]bool {
	out := map[string]bool{}
	for _, it := range items {
		out[it.Name] = true
	}
	return out
}

func TestTiersNest(t *testing.T) {
	server := names(ForTier(Server))
	daily := names(ForTier(Daily))
	ultimate := names(ForTier(Ultimate))

	for name := range server {
		if !daily[name] || !ultimate[name] {
			t.Errorf("server item %q missing from a higher tier", name)
		}
	}
	for name := range daily {
		if !ultimate[name] {
			t.Errorf("daily item %q missing from ultimate", name)
		}
	}
	if len(server) >= len(daily) || len(daily) >= len(ultimate) {
		t.Errorf("tier sizes should strictly grow: server=%d daily=%d ultimate=%d",
			len(server), len(daily), len(ultimate))
	}
}

func TestResolveIncludeExclude(t *testing.T) {
	items, err := Resolve(Server, []string{"discord"}, []string{"git", "tmux"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(items)
	if !got["discord"] {
		t.Error("include discord was dropped")
	}
	if got["git"] || got["tmux"] {
		t.Error("excluded items survived")
	}
	if !got["fzf"] {
		t.Error("untouched tier item fzf missing")
	}
}

func TestResolveUnknownNames(t *testing.T) {
	for _, args := range [][2][]string{
		{{"nonsense"}, nil},
		{nil, {"nonsense"}},
	} {
		if _, err := Resolve(Daily, args[0], args[1]); err == nil {
			t.Errorf("expected error for unknown name in include=%v exclude=%v", args[0], args[1])
		} else if !strings.Contains(err.Error(), "nonsense") {
			t.Errorf("error should name the unknown item, got: %v", err)
		}
	}
}

func TestBinary(t *testing.T) {
	cases := []struct {
		item string
		want string
	}{
		{"neovim", "nvim"},
		{"ripgrep", "rg"},
		{"git-delta", "delta"},
		{"claude-code", "claude"},
		{"git", "git"},
		{"arc", ""},
		{"meslo-nerd-font", ""},
	}
	for _, c := range cases {
		it, ok := ByName(c.item)
		if !ok {
			t.Fatalf("catalog item %q missing", c.item)
		}
		if got := it.Binary(); got != c.want {
			t.Errorf("%s.Binary() = %q, want %q", c.item, got, c.want)
		}
	}
}

func TestCatalogInvariants(t *testing.T) {
	seen := map[string]bool{}
	for _, it := range Items {
		if seen[it.Name] {
			t.Errorf("duplicate catalog name %q", it.Name)
		}
		seen[it.Name] = true
		if it.Brew == "" {
			t.Errorf("%q has no brew name", it.Name)
		}
		if it.Cask && (it.Apt != "" || it.Dnf != "") {
			t.Errorf("cask %q claims a Linux package", it.Name)
		}
	}
}
