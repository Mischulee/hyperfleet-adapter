# Sourced by the example dry-run scripts with their [create|delete] argument.
# Changes to the repository root, builds the adapter, and defines dryrun_trace.

dryrun_mode="${1:-create}"
case "$dryrun_mode" in
  create) dryrun_responses=test/testdata/dryrun/dryrun-api-responses.json ;;
  delete) dryrun_responses=test/testdata/dryrun/dryrun-delete-api-responses.json ;;
  *) echo "Usage: $0 [create|delete]" >&2; exit 2 ;;
esac

cd "$(git rev-parse --show-toplevel)"
# A normal build, so the examples also pass the adapter version check.
make build

# dryrun_trace <adapter-config> <task-config> <discovery> prints the JSON trace.
dryrun_trace() {
  HYPERFLEET_TRACING_ENABLED=false bin/hyperfleet-adapter serve \
    --config "$1" \
    --task-config "$2" \
    --dry-run-event test/testdata/dryrun/event.json \
    --dry-run-api-responses "$dryrun_responses" \
    --dry-run-discovery "$3" \
    --dry-run-verbose \
    --dry-run-output json
}
