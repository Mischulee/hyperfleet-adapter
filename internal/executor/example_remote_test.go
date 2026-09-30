package executor_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/configloader"
	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/desireclient"
	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/desireclient/desiretest"
	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/executor"
	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/hyperfleetapi"
	"github.com/openshift-hyperfleet/hyperfleet-adapter/internal/transportclient"
	"github.com/openshift-hyperfleet/hyperfleet-applier/pkg/desire"
	"github.com/openshift-hyperfleet/hyperfleet-applier/pkg/desire/store/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemoteTwoResourceExampleLifecycle uses the shipped task and the real
// memory-backed remote client. The CLI dry run uses a recorder instead.
func TestRemoteTwoResourceExampleLifecycle(t *testing.T) {
	ctx := t.Context()
	root := filepath.Join("..", "..", "charts", "examples", "remote-two-resources")
	config, err := configloader.LoadConfig(
		configloader.WithAdapterConfigPath(filepath.Join(root, "adapter-config.yaml")),
		configloader.WithTaskConfigPath(filepath.Join(root, "adapter-task-config.yaml")),
	)
	require.NoError(t, err)

	const owner = "hyperfleet-adapter"
	store := memory.New()
	api := hyperfleetapi.NewMockClient()
	adapterExecutor, err := executor.NewBuilder().
		WithConfig(config).
		WithAPIClient(api).
		WithTransportRegistry(transportclient.Registry{
			"remote-primary": desireclient.NewClient(store, owner),
		}).
		Build()
	require.NoError(t, err)

	namespace := desiretest.TestIdentity{
		ManagementCluster: "abc123", Resource: "namespaces", Name: "abc123-remote",
	}
	configMap := desiretest.TestIdentity{
		ManagementCluster: "abc123", Resource: "configmaps",
		Namespace: "abc123-remote", Name: "cluster-config",
	}
	desiretest.PutUnsyncedReadDesire(t, ctx, store, namespace.Read(), owner)
	desiretest.PutUnsyncedReadDesire(t, ctx, store, configMap.Read(), owner)

	run := func(deleting bool, available, finalized string) *executor.ExecutionResult {
		t.Helper()
		api.Reset()
		status := `{"generation":"77"}`
		if deleting {
			status = `{"generation":"77","deleted_time":"2026-09-29T00:00:00Z"}`
		}
		api.GetResponse = &hyperfleetapi.Response{StatusCode: 200, Body: []byte(status)}
		result := adapterExecutor.Execute(ctx, map[string]any{"id": "abc123", "kind": "Cluster"})
		require.Equal(t, executor.StatusSuccess, result.Status, "errors=%v", result.Errors)
		require.NotNil(t, result.ExecutionContext)
		require.Len(t, result.PostActionResults, 1)
		require.True(t, result.PostActionResults[0].APICallMade)
		request := api.GetLastRequest()
		require.NotNil(t, request)
		require.Equal(t, "PUT", request.Method)
		var payload struct {
			Conditions []struct {
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"conditions"`
		}
		require.NoError(t, json.Unmarshal(request.Body, &payload))
		conditions := make(map[string]string, len(payload.Conditions))
		for _, condition := range payload.Conditions {
			conditions[condition.Type] = condition.Status
		}
		assert.Equal(t, available, conditions["Available"])
		assert.Equal(t, finalized, conditions["Finalized"])
		assert.Equal(t, "True", conditions["Health"])
		return result
	}

	// Unsynced and stale mirrors are not evidence of failure, so Available stays Unknown.
	result := run(false, "Unknown", "False")
	assert.Equal(t, executor.ResourceStateUnsynced, result.ExecutionContext.ResourceStates["namespace"])
	assert.Equal(t, executor.ResourceStateUnsynced, result.ExecutionContext.ResourceStates["configMap"])

	desiretest.MarkReadDesireSynced(t, ctx, store, namespace.Read(), []byte(`{
		"apiVersion":"v1","kind":"Namespace",
		"metadata":{"name":"abc123-remote","annotations":{"hyperfleet.io/generation":"77"}},
		"status":{"phase":"Active"}
	}`))
	desiretest.MarkReadDesireNotFound(t, ctx, store, configMap.Read())
	run(false, "Unknown", "False")

	desiretest.MarkReadDesireSynced(t, ctx, store, configMap.Read(), []byte(`{
		"apiVersion":"v1","kind":"ConfigMap",
		"metadata":{"name":"cluster-config","namespace":"abc123-remote",
			"annotations":{"hyperfleet.io/generation":"76"}}
	}`))
	run(false, "Unknown", "False")

	desiretest.MarkReadDesireSynced(t, ctx, store, configMap.Read(), []byte(`{
		"apiVersion":"v1","kind":"ConfigMap",
		"metadata":{"name":"cluster-config","namespace":"abc123-remote",
			"annotations":{"hyperfleet.io/generation":"77"}}
	}`))
	run(false, "True", "False")

	result = run(true, "True", "False")
	assert.NotEqual(t, executor.ResourceStateConfirmedDeleted, result.ExecutionContext.ResourceStates["configMap"])
	_, err = store.GetDeleteDesire(ctx, configMap.Delete())
	require.NoError(t, err)
	_, err = store.GetDeleteDesire(ctx, namespace.Delete())
	require.ErrorIs(t, err, desire.ErrNotFound, "the parent must wait for the child")

	desiretest.MarkDeleteDesireConfirmed(t, ctx, store, configMap.Delete())
	result = run(true, "False", "False")
	assert.Equal(t, executor.ResourceStateConfirmedDeleted, result.ExecutionContext.ResourceStates["configMap"])
	_, err = store.GetDeleteDesire(ctx, namespace.Delete())
	require.ErrorIs(t, err, desire.ErrNotFound, "the parent ran before the child confirmation in this event")

	result = run(true, "False", "False")
	assert.Equal(t, executor.ResourceStateConfirmedDeleted, result.ExecutionContext.ResourceStates["configMap"])
	_, err = store.GetDeleteDesire(ctx, namespace.Delete())
	require.NoError(t, err, "the parent should now be requested for deletion")

	desiretest.MarkDeleteDesireConfirmed(t, ctx, store, namespace.Delete())
	result = run(true, "False", "True")
	assert.Equal(t, executor.ResourceStateConfirmedDeleted, result.ExecutionContext.ResourceStates["namespace"])
	assert.Equal(t, executor.ResourceStateConfirmedDeleted, result.ExecutionContext.ResourceStates["configMap"])
	_, err = store.GetDeleteDesire(ctx, namespace.Delete())
	assert.ErrorIs(t, err, desire.ErrNotFound)
	_, err = store.GetDeleteDesire(ctx, configMap.Delete())
	assert.ErrorIs(t, err, desire.ErrNotFound)
}
