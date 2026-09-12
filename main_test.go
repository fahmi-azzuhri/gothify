package main

import (
	"strings"
	"testing"
)

func TestOptionsFromFlags(t *testing.T) {
	options, err := optionsFromFlags("demo", "chi", "templ", "tailwind", "session", "example.com/demo", true, true)
	if err != nil {
		t.Fatal(err)
	}
	if options.Router != "Chi (lightweight and idiomatic Go)" || !options.PWA || !options.Embed {
		t.Fatalf("unexpected options: %+v", options)
	}
}

func TestOptionsFromFlagsRejectsUnknownValue(t *testing.T) {
	if _, err := optionsFromFlags("demo", "mux", "templ", "plain", "session", "", false, false); err == nil {
		t.Fatal("expected invalid router error")
	}
}

func TestFilesForIncludesSelectedFeatures(t *testing.T) {
	options, err := optionsFromFlags("demo", "fiber", "inertia", "tailwind", "jwt", "example.com/demo", true, true)
	if err != nil {
		t.Fatal(err)
	}
	files := filesFor(options)
	paths := make(map[string]string, len(files))
	for _, file := range files {
		paths[file.Path] = file.Body
	}
	if !strings.Contains(paths["cmd/server/main.go"], "fiber.New()") {
		t.Error("fiber server template missing")
	}
	if !strings.Contains(paths["internal/auth/auth.go"], "IssueRefreshToken") {
		t.Error("JWT auth template missing")
	}
	if _, ok := paths["web/assets.go"]; !ok {
		t.Error("embedded assets template missing")
	}
	if _, ok := paths["package.json"]; !ok {
		t.Error("frontend package manifest missing")
	}
}
