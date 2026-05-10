package core

import "testing"

func TestFlattenVersions(t *testing.T) {
	got := flattenVersions(map[string][]string{
		"1.21":   []string{"1.21.1", "1.21.2"},
		"latest": []string{"1.21.2"},
	})
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestSelectPaperBuild(t *testing.T) {
	build, err := selectPaperBuild([]paperBuild{
		{ID: 1, Channel: "EXPERIMENTAL"},
		{ID: 2, Channel: "STABLE"},
		{ID: 3, Channel: "STABLE"},
	}, "latest", "STABLE")
	if err != nil {
		t.Fatal(err)
	}
	if build.ID != 3 {
		t.Fatalf("got build %d", build.ID)
	}
}
