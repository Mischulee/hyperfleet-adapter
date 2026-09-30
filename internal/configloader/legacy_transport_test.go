package configloader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const legacyAdapterYAML = `
adapter:
  name: test-adapter
  version: "0.3.0"
`

const legacyMaestroAdapterYAML = legacyAdapterYAML + `
clients:
  maestro:
    grpc_server_address: maestro-grpc:8090
    http_server_address: http://maestro:8000
    source_id: test-adapter
`

func legacyTask(schemaVersion, transport string) string {
	task := ""
	if schemaVersion != "" {
		task = "schema_version: \"" + schemaVersion + "\"\n"
	}
	return task + `
params:
  - name: clusterId
    source: event.id
resources:
  - name: work
    transport:
` + transport + `
    manifest:
      apiVersion: work.open-cluster-management.io/v1
      kind: ManifestWork
      metadata: {name: test, namespace: "{{ .clusterId }}"}
    discovery: {by_name: test, namespace: "{{ .clusterId }}"}
`
}

const legacyMaestroTransport = `      client: maestro
      maestro:
        target_cluster: "{{ .clusterId }}"`

func TestLoadLegacyTransportForm(t *testing.T) {
	t.Run("kubernetes selects the local transport", func(t *testing.T) {
		config, err := loadRemoteConfig(t, legacyAdapterYAML, legacyTask("", `      client: kubernetes`))
		require.NoError(t, err)
		assert.Equal(t, TransportClientKubernetes, config.Resources[0].GetTransportName())
		assert.False(t, config.Resources[0].IsMaestroTransport())
	})

	t.Run("a Maestro client may also serve local resources", func(t *testing.T) {
		config, err := loadRemoteConfig(t, legacyMaestroAdapterYAML, legacyTask("", `      client: kubernetes`))
		require.NoError(t, err)
		assert.False(t, UsesMaestro(config.Resources))
	})

	t.Run("maestro selects Maestro with its target cluster", func(t *testing.T) {
		config, err := loadRemoteConfig(t, legacyMaestroAdapterYAML, legacyTask("", legacyMaestroTransport))
		require.NoError(t, err)
		require.True(t, config.Resources[0].IsMaestroTransport())
		assert.Equal(t, "{{ .clusterId }}", config.Resources[0].Transport.Maestro.TargetCluster)
		assert.True(t, UsesMaestro(config.Resources))
	})

	for name, tc := range map[string]struct {
		adapter, task, want string
	}{
		"rejected with schema_version 2.0": {
			legacyMaestroAdapterYAML, legacyTask("2.0", legacyMaestroTransport),
			`resources[0].transport: the object form is not supported with schema_version "2.0"`,
		},
		"maestro needs clients.maestro": {
			legacyAdapterYAML, legacyTask("", legacyMaestroTransport),
			`resources[0].transport selects "maestro", but clients.maestro is not configured`,
		},
		"maestro needs its settings": {
			legacyMaestroAdapterYAML, legacyTask("", `      client: maestro`),
			`resources[0].transport.maestro is required for client "maestro"`,
		},
		"maestro needs target_cluster": {
			legacyMaestroAdapterYAML, legacyTask("", `      client: maestro
      maestro: {}`),
			"resources[0].transport.maestro.target_cluster is required",
		},
		"target_cluster variables must be defined": {
			legacyMaestroAdapterYAML, legacyTask("", `      client: maestro
      maestro:
        target_cluster: "{{ .missing }}"`),
			`target_cluster uses undefined variable "missing"`,
		},
		"unknown client": {
			legacyAdapterYAML, legacyTask("", `      client: remote-primary`),
			`resources[0].transport.client must be "kubernetes" or "maestro"`,
		},
		"maestro settings on kubernetes": {
			legacyAdapterYAML, legacyTask("", `      client: kubernetes
      maestro:
        target_cluster: cluster-1`),
			`resources[0].transport.maestro is only valid with client "maestro"`,
		},
		"unknown legacy field": {
			legacyAdapterYAML, legacyTask("", `      client: kubernetes
      desire: {}`),
			"field desire not found",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := loadRemoteConfig(t, tc.adapter, tc.task)
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestResourceTransportYAMLRoundTrip(t *testing.T) {
	for name, input := range map[string]string{
		"name":   "transport: remote-primary\n",
		"legacy": "transport:\n    maestro:\n        target_cluster: cluster-1\n    client: maestro\n",
	} {
		t.Run(name, func(t *testing.T) {
			var resource struct {
				Transport *ResourceTransport `yaml:"transport"`
			}
			require.NoError(t, yaml.Unmarshal([]byte(input), &resource))
			out, err := yaml.Marshal(resource)
			require.NoError(t, err)
			assert.Equal(t, input, string(out))
		})
	}

	var resource struct {
		Transport *ResourceTransport `yaml:"transport"`
	}
	require.ErrorContains(t, yaml.Unmarshal([]byte("transport: [a]\n"), &resource), "transport must be a transport name")
}
