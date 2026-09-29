package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractGVKFromString(t *testing.T) {
	tests := []struct {
		name        string
		manifest    string
		wantGroup   string
		wantVersion string
		wantKind    string
		wantEmpty   bool
	}{
		{
			name:        "simple v1 ConfigMap",
			manifest:    "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n",
			wantVersion: "v1",
			wantKind:    "ConfigMap",
		},
		{
			name:        "with Go template value expressions",
			manifest:    "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: \"{{ .clusterId }}\"\n",
			wantVersion: "v1",
			wantKind:    "ConfigMap",
		},
		{
			name: "with structural Go template directives",
			manifest: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n" +
				"  name: \"test-{{ .clusterId }}\"\n  labels:\n    app: test\n" +
				"{{ if .testRunId }}\n    run-id: \"{{ .testRunId }}\"\n{{ end }}\n" +
				"data:\n  key: value\n",
			wantVersion: "v1",
			wantKind:    "ConfigMap",
		},
		{
			name:        "apps/v1 Deployment",
			manifest:    "apiVersion: apps/v1\nkind: Deployment\n",
			wantGroup:   "apps",
			wantVersion: "v1",
			wantKind:    "Deployment",
		},
		{
			name:      "missing kind",
			manifest:  "apiVersion: v1\nmetadata:\n  name: test\n",
			wantEmpty: true,
		},
		{
			name:      "missing apiVersion",
			manifest:  "kind: ConfigMap\nmetadata:\n  name: test\n",
			wantEmpty: true,
		},
		{
			name:      "empty string",
			manifest:  "",
			wantEmpty: true,
		},
		{
			name: "ignores nested apiVersion and kind",
			manifest: "metadata:\n  name: test\n  ownerReferences:\n    - apiVersion: v1\n" +
				"      kind: Cluster\n      name: owner\napiVersion: v1\nkind: ConfigMap\n",
			wantVersion: "v1",
			wantKind:    "ConfigMap",
		},
		{
			name:        "quoted values with trailing comments",
			manifest:    "apiVersion: \"apps/v1\" # group\nkind: 'Deployment' # kind\n",
			wantGroup:   "apps",
			wantVersion: "v1",
			wantKind:    "Deployment",
		},
		{
			name:        "same kind in both template branches",
			manifest:    "apiVersion: v1\n{{ if .x }}\nkind: ConfigMap\n{{ else }}\nkind: ConfigMap\n{{ end }}\n",
			wantVersion: "v1",
			wantKind:    "ConfigMap",
		},
		{
			name:      "different kinds in template branches",
			manifest:  "apiVersion: v1\n{{ if .x }}\nkind: ConfigMap\n{{ else }}\nkind: Secret\n{{ end }}\n",
			wantEmpty: true,
		},
		{
			name:      "templated kind",
			manifest:  "apiVersion: v1\nkind: \"{{ .kind }}\"\n",
			wantEmpty: true,
		},
		{
			name:      "only nested kind",
			manifest:  "apiVersion: v1\nmetadata:\n  kind: ConfigMap\n",
			wantEmpty: true,
		},
		{
			name:      "invalid content",
			manifest:  "not: valid: yaml: {{{}",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gvk := ExtractGVKFromString(tt.manifest)

			if tt.wantEmpty {
				assert.True(t, gvk.Empty(), "expected empty GVK")
			} else {
				assert.Equal(t, tt.wantGroup, gvk.Group)
				assert.Equal(t, tt.wantVersion, gvk.Version)
				assert.Equal(t, tt.wantKind, gvk.Kind)
			}
		})
	}
}
