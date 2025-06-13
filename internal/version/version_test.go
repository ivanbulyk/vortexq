package version

import "testing"

// Test NewVersion returns the package variables
func TestNewVersion(t *testing.T) {
	v := NewVersion()
	if v.Project != Project {
		t.Errorf("Project = %q; want %q", v.Project, Project)
	}
	if v.BuildTime != BuildTime {
		t.Errorf("BuildTime = %q; want %q", v.BuildTime, BuildTime)
	}
	if v.Commit != Commit {
		t.Errorf("Commit = %q; want %q", v.Commit, Commit)
	}
	if v.Release != Release {
		t.Errorf("Release = %q; want %q", v.Release, Release)
	}
}
