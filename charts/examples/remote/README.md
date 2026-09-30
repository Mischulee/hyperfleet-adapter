# Remote ConfigMap example

This v2 task applies a plain ConfigMap through the named `remote-primary`
transport. The task uses `schema_version: "2.0"` and a relative manifest
reference, so the task and ConfigMap can be copied together.

`adapter-config.yaml` is the installable configuration. Replace its Redis URL
and provision a shared Redis store and remote applier before deployment. The
CLI dry run uses the same file but records operations with a mock transport,
so it never connects to Redis.

Run `bash charts/examples/remote/dryrun.sh create` or `delete` from the
repository root. The script builds the adapter and checks the JSON trace. Deletion needs a later reconciliation before confirmation; an
absent or stale mirror alone is not proof of deletion.

The chart overlay is `charts/examples/remote/values.yaml`. See the
[authoring guide](../../../docs/adapter-authoring-guide.md) for task concepts.
