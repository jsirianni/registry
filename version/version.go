package version

// injected at build time
var (
	version string
	gitHash string
	date    string
)

// Version contains fields with build information
type Version struct {
	// The version tag
	Version string `json:"version"`

	// The full commit hash
	CommitHash string `json:"commit_hash"`

	// The build date
	BuildDate string `json:"build_date"`
}

// BuildVersion returns the release version
func BuildVersion() Version {
	return Version{
		Version:    version,
		CommitHash: gitHash,
		BuildDate:  date,
	}
}
