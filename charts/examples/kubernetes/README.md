# Local Namespace and Job example

This v2 task applies a Namespace and then a hello-world Job through the
implicit local Kubernetes transport. The Namespace manifest is inline;
`job.yaml` is a relative reference mounted beside the task config by the
chart overlay.

The task reads the cluster generation and deletion state from the HyperFleet
API. Its `Available` condition reads the live Job `Complete` condition, and
`Finalized` waits for confirmed deletion of both resources on a deletion
event. Resource execution follows list order; remote delivery is illustrated
in the [remote examples](../README.md).

Render the overlay with:

```bash
helm template example charts -f charts/examples/kubernetes/values.yaml \
  --set image.registry=quay.io \
  --set image.repository=openshift-hyperfleet/hyperfleet-adapter \
  --set image.tag=test
```

Test the local mock path with `bash charts/examples/kubernetes/dryrun.sh`.
Replace the broker and image placeholders in `values.yaml` before installing.
See the [authoring guide](../../../docs/adapter-authoring-guide.md).
