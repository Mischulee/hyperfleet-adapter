# Adapter examples

Every task example uses `schema_version: "2.0"` and plain Kubernetes
manifests. A resource without a transport uses the local Kubernetes client.

| Example | Delivery | Resources |
|---|---|---|
| [kubernetes](./kubernetes/) | Local | Namespace and Job |
| [remote](./remote/) | Named remote transport | ConfigMap |
| [remote-two-resources](./remote-two-resources/) | One named remote transport | Namespace, then ConfigMap |

The remote examples use a shared Redis store in their installable deployment
configs. Supply a Redis service and remote applier separately. The CLI dry run
records operations with a mock transport and never connects to Redis.

Render an overlay with:

```bash
helm template example charts -f charts/examples/remote/values.yaml \
  --set image.registry=quay.io \
  --set image.repository=openshift-hyperfleet/hyperfleet-adapter \
  --set image.tag=test
```

Replace broker and image placeholders before installing. See the
[authoring guide](../../docs/adapter-authoring-guide.md) for the v2 model.
