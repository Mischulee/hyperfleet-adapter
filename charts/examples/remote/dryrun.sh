#!/usr/bin/env bash
set -euo pipefail

source "$(git rev-parse --show-toplevel)/charts/examples/dryrun-lib.sh"
dryrun_trace \
  charts/examples/remote/adapter-config.yaml \
  charts/examples/remote/adapter-task-config.yaml \
  charts/examples/remote/dryrun-discovery.json |
  jq --arg mode "$dryrun_mode" -e '
  [.apiRequests[] | select(.method == "PUT") | .requestBody | fromjson][0] as $body
  | .status == "success"
    and .discoveredResources.configMap.kind == "ConfigMap"
    and (any(.transportOperations[]; .kind == "ConfigMap" and .targetCluster == "abc123" and .targetResource == "configmaps"))
    and ($body.conditions | any(.[]; .type == "Health" and .status == "True"))
    and ($body.conditions | any(.[]; .type == "Finalized" and .status == "False"))
    and (if $mode == "create" then
      any(.transportOperations[]; .operation == "apply" and .kind == "ConfigMap")
      and ($body.conditions | any(.[]; .type == "Available" and .status == "True"))
    else
      any(.transportOperations[]; .operation == "delete" and .kind == "ConfigMap")
      and ($body.conditions | any(.[]; .type == "Available" and .status == "Unknown"))
      and ($body.conditions | any(.[]; .type == "Finalized" and .reason == "CleanupInProgress"))
    end)
'
