package main

import (
	"context"
	"testing"

	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/configloader"
	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/dryrun"
	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/executor"
	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/transportregistry"
	"github.com/stretchr/testify/require"
)

func TestDryRunLogOptionsDefaults(t *testing.T) {
	level, format := buildDryRunLogOptions()

	require.Equal(t, "warn", level)
	require.Equal(t, "text", format)
}

func TestDryRunLogOptionsHonorsLevelOverride(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")

	level, _ := buildDryRunLogOptions()

	require.Equal(t, "debug", level)
}

func TestLogOptionsFlagOverridesEnv(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	prev := logLevel
	logLevel = "error"
	t.Cleanup(func() { logLevel = prev })

	level, _, _ := buildLogOptions(nil)

	require.Equal(t, "error", level, "CLI flag must take precedence over LOG_LEVEL")
}

func TestLogOptionsBootstrapDefaults(t *testing.T) {
	prevLevel, prevFormat, prevOutput := logLevel, logFormat, logOutput
	logLevel, logFormat, logOutput = "", "", ""
	t.Cleanup(func() {
		logLevel, logFormat, logOutput = prevLevel, prevFormat, prevOutput
	})

	level, format, output := buildLogOptions(nil)

	require.Empty(t, level, "bootstrap level is passed to hfl.ParseLevel")
	require.Empty(t, format, "bootstrap format is passed to hfl.ParseFormat")
	require.Empty(t, output, "bootstrap output is passed to hfl.ParseOutput")
}

func TestBuildExecutor_DryRunNamedRemoteTransport(t *testing.T) {
	config := &configloader.Config{
		Adapter: configloader.AdapterInfo{Name: "test-adapter"},
		Stores: map[string]configloader.StoreDefinition{
			"desired-memory": {Type: configloader.StoreTypeMemory},
		},
		Transports: map[string]configloader.TransportDefinition{
			"remote-primary": {
				Type: configloader.TransportTypeRemote, Store: "desired-memory", TargetCluster: "cluster-1",
				ResourcePlurals: map[string]string{"v1/ConfigMap": "configmaps"},
			},
		},
		Resources: []configloader.Resource{{
			Name:      "test-resource",
			Transport: new("remote-primary"),
			Manifest: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]interface{}{"name": "test-config", "namespace": "default"},
			},
		}},
	}
	recorder := dryrun.NewDryrunTransportClient()
	runtime, err := transportregistry.BuildRecording(config, recorder)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, runtime.Close()) })
	apiClient, err := dryrun.NewDryrunAPIClient(nil)
	require.NoError(t, err)

	exec, err := buildExecutor(config, apiClient, runtime.Registry, nil)
	require.NoError(t, err)
	result := exec.Execute(context.Background(), map[string]interface{}{"id": "cluster-1", "kind": "Cluster"})

	require.Equal(t, executor.StatusSuccess, result.Status)
	require.Len(t, recorder.Records, 1)
	require.Equal(t, "apply", recorder.Records[0].Operation)
}
