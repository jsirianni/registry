package store

import (
	"testing"

	"github.com/jsirianni/registry/model"
	"github.com/stretchr/testify/require"
)

func TestMapStore(t *testing.T) {
	m := NewMapStore()
	require.Empty(t, m.providers)

	output := m.Read("test")
	require.Nil(t, output)

	input := model.ProviderVersion{
		Version:   "5.1.0",
		Protocols: []string{"10.1"},
	}

	err := m.Write("test", model.ProviderVersions{
		Versions: []model.ProviderVersion{input},
	})
	require.NoError(t, err)

	output = m.Read("test")
	require.NotNil(t, output)
	require.Equal(t, "5.1.0", output.Versions[0].Version)
	require.Equal(t, []string{"10.1"}, output.Versions[0].Protocols)
}
