package modrinth

import "testing"

func TestSelectFilePrefersPrimaryJar(t *testing.T) {
	version, file, err := SelectFile([]Version{{
		VersionNumber: "1.0.0",
		Files: []File{
			{Filename: "sources.jar", Primary: false},
			{Filename: "plugin.jar", Primary: true},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if version.VersionNumber != "1.0.0" || file.Filename != "plugin.jar" {
		t.Fatalf("got %s %s", version.VersionNumber, file.Filename)
	}
}

func TestBuildFacets(t *testing.T) {
	facets := buildFacets("mod", "fabric", "1.21.1")
	if len(facets) != 3 {
		t.Fatalf("got %#v", facets)
	}
}
