package k8sclient

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GVKFromKindAndAPIVersion creates a GroupVersionKind from kind and apiVersion
// strings, for example "Deployment" and "apps/v1". The executor reads a
// resource's GVK from its manifest with configloader.Resource.StaticGVK instead.
func GVKFromKindAndAPIVersion(kind, apiVersion string) (schema.GroupVersionKind, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	return schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    kind,
	}, nil
}

// GVKFromUnstructured extracts GroupVersionKind from an unstructured object.
// It returns an empty GVK for a nil object.
func GVKFromUnstructured(obj *unstructured.Unstructured) schema.GroupVersionKind {
	if obj == nil {
		return schema.GroupVersionKind{}
	}
	return obj.GroupVersionKind()
}

// NOTE: CommonResourceKinds has been moved to test_helpers.go
// It is for testing purposes only and should not be used in production code.
