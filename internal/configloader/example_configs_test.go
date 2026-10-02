package configloader

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShippedV2ConfigsLoad(t *testing.T) {
	root := filepath.Join("..", "..")
	cases := []struct {
		name        string
		adapterPath string
		taskPath    string
		resources   int
	}{
		{
			name: "template", adapterPath: "configs/adapter-config-template.yaml",
			taskPath: "configs/adapter-task-config-template.yaml", resources: 1,
		},
		{
			name: "kubernetes", adapterPath: "charts/examples/kubernetes/adapter-config.yaml",
			taskPath: "charts/examples/kubernetes/adapter-task-config.yaml", resources: 2,
		},
		{
			name: "two resources", adapterPath: "charts/examples/remote-two-resources/adapter-config.yaml",
			taskPath: "charts/examples/remote-two-resources/adapter-task-config.yaml", resources: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := LoadConfig(
				WithAdapterConfigPath(filepath.Join(root, tc.adapterPath)),
				WithTaskConfigPath(filepath.Join(root, tc.taskPath)),
			)
			require.NoError(t, err)
			require.Equal(t, "2.0", config.SchemaVersion)
			require.Len(t, config.Resources, tc.resources)
			for _, resource := range config.Resources {
				require.NotNil(t, resource.Manifest, "resource %q manifest", resource.Name)
			}
		})
	}
}
