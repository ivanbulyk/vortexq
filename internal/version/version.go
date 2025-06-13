package version

type Version struct {
	// Project is a name of the project
	Project string `json:"project"`
	// BuildTime is a time label of the moment when the binary was built
	BuildTime string `json:"build_time"`
	// Commit is a last commit hash at the moment when the binary was built
	Commit string `json:"commit"`
	// Release is a semantic version of the current build
	Release string `json:"release"`
}

// NewVersion creates a new Version instance with the current build information.
func NewVersion() *Version {
	return &Version{
		Project:   Project,
		BuildTime: BuildTime,
		Commit:    Commit,
		Release:   Release,
	}
}

var (
	// Project is a name of the project
	Project = "vortexq"
	// BuildTime is a time label of the moment when the binary was built
	BuildTime = "unset"
	// Commit is a last commit hash at the moment when the binary was built
	Commit = "unset"
	// Release is a semantic version of the current build
	Release = "unset"
)
