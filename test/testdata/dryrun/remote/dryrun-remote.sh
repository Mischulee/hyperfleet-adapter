#!/usr/bin/env bash
set -euo pipefail

source "$(git rev-parse --show-toplevel)/charts/examples/dryrun-lib.sh"
dryrun_trace \
  test/testdata/dryrun/remote/dryrun-remote-adapter-config.yaml \
  test/testdata/dryrun/remote/dryrun-remote-task-config.yaml \
  test/testdata/dryrun/remote/dryrun-remote-discovery.json |
  jq --arg mode "$dryrun_mode" -e '
  .status == "success"
  and .discoveredResources.cluster_config.kind == "ConfigMap"
  and any(.transportOperations[];
    .kind == "ConfigMap" and .operation == (if $mode == "create" then "apply" else "delete" end)
    and .targetCluster == "abc123" and .targetResource == "configmaps")
'
