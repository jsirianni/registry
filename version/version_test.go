package version

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildVersion(t *testing.T) {
	// without ldflags, version should be empty strings
	expect := Version{}
	output := BuildVersion()
	require.Equal(t, expect, output)

	version = "v2.1.1"
	gitHash = "101010"
	d := time.Now().String()
	date = d
	defer func() {
		version = ""
		gitHash = ""
		date = ""
	}()

	expect = Version{
		Version:    "v2.1.1",
		CommitHash: "101010",
		BuildDate:  d,
	}
	output = BuildVersion()
	require.Equal(t, expect, output)
}

// Do not set default values
func Test_defaultValues(t *testing.T) {
	require.Equal(t, "", version)
	require.Equal(t, "", gitHash)
	require.Equal(t, "", date)
}
