# Remote Namespace and ConfigMap example

This v2 task uses one named remote transport for two plain Kubernetes
resources. The Namespace is first in the list, so it is applied before the
ConfigMap. Both manifests carry the requested HyperFleet generation. Readiness
uses the live Namespace `status.phase` and generation annotations on both
objects.

During deletion the ConfigMap is requested first. The Namespace delete gate
opens only when `resource_states.configMap` is `confirmed_deleted`; final
status waits for both resources to reach that state. Remote reads can take
several events to converge. A missing mirror is not a deletion confirmation.

`adapter-config.yaml` requires a separately provisioned shared Redis store
and remote applier. Before deployment, create a Secret with a `url` key
containing the authenticated `rediss://` Redis URL, and replace
`CHANGE_ME-redis` in `values.yaml` with that Secret's name. The chart injects
the URL through `HYPERFLEET_STORES_REMOTE_STORE_URL`; the adapter config
ConfigMap contains only a placeholder. The CLI dry run records operations
with a mock transport and never connects to Redis.

Run `bash charts/examples/remote-two-resources/dryrun.sh create` or
`delete` from the repository root. The chart overlay is
`charts/examples/remote-two-resources/values.yaml`. See the
[authoring guide](../../../docs/adapter-authoring-guide.md).
