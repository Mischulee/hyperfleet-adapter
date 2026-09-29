package manifest

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ExtractGVKFromString extracts apiVersion and kind from a YAML string
// by scanning lines. This handles manifests with Go template directives
// that would fail full YAML parsing.
//
// Only top-level keys (no indentation) count, so a nested apiVersion or kind,
// such as one in metadata.ownerReferences, cannot be mistaken for the
// object's own. apiVersion and kind must be static: a templated value, or a
// key that appears more than once with different values (for example in
// {{ if }}/{{ else }} branches), yields an empty GVK. This function is also
// used in deletion flows where no rendered manifest is available
// (discover → delete, no render/apply).
func ExtractGVKFromString(manifest string) schema.GroupVersionKind {
	var apiVersion, kind string
	for line := range strings.SplitSeq(manifest, "\n") {
		var target *string
		value, ok := strings.CutPrefix(line, "apiVersion:")
		if ok {
			target = &apiVersion
		} else if value, ok = strings.CutPrefix(line, "kind:"); ok {
			target = &kind
		} else {
			continue
		}
		value = topLevelScalar(value)
		if value == "" || strings.Contains(value, "{{") || (*target != "" && *target != value) {
			return schema.GroupVersionKind{}
		}
		*target = value
	}
	if apiVersion == "" || kind == "" {
		return schema.GroupVersionKind{}
	}
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return schema.GroupVersionKind{}
	}
	return gv.WithKind(kind)
}

// topLevelScalar strips a trailing comment, surrounding whitespace, and quotes
// from a plain or quoted single-line YAML scalar.
func topLevelScalar(value string) string {
	value, _, _ = strings.Cut(value, " #")
	return strings.Trim(strings.TrimSpace(value), "\"'")
}
