#!/usr/bin/env bash
set -euo pipefail

source "$(git rev-parse --show-toplevel)/charts/examples/dryrun-lib.sh"
dryrun_trace \
  charts/examples/kubernetes/adapter-config.yaml \
  charts/examples/kubernetes/adapter-task-config.yaml \
  charts/examples/kubernetes/dryrun-discovery.json |
  jq --arg mode "$dryrun_mode" -e '
  [.apiRequests[] | select(.method == "PUT") | .requestBody | fromjson][0] as $body
  | .status == "success"
    and (all(.transportOperations[]; (.targetCluster // "") == ""))
    and (if $mode == "create" then
      any(.transportOperations[]; .operation == "apply" and .kind == "Namespace")
      and any(.transportOperations[]; .operation == "apply" and .kind == "Job")
      and .discoveredResources.helloWorldJob.kind == "Job"
      and ($body.conditions | any(.[]; .type == "Available" and .status == "True"))
    else
      any(.transportOperations[]; .operation == "delete" and .kind == "Namespace")
      and any(.transportOperations[]; .operation == "delete" and .kind == "Job")
      and ($body.conditions | any(.[]; .type == "Finalized" and .status == "True"))
    end)
'
