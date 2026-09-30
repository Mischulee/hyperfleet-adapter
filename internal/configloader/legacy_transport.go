package configloader

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// TODO(HYPERFLEET-1504): delete this file, and every other TODO(HYPERFLEET-1504)
// spot, once the Maestro cutover (HYPERFLEET-1503, 1505, 1506) is done. It is a
// transitional shim that keeps deployed pre-v2 task configs, including Maestro
// ones, loading so adapter main stays deployable and e2e stays green meanwhile.

const schemaVersionV2 = "2.0"

// ResourceTransport is a resource's transport reference: a named deployment
// route, or the pre-v2 object form {client, maestro}. A schema_version "2.0"
// task must name a transport.
type ResourceTransport struct {
	// Maestro contains maestro-specific transport settings (object form with client "maestro")
	Maestro *MaestroTransportConfig `yaml:"maestro,omitempty"`
	// Name is the route name, or the object form's client.
	Name string
	// Legacy is true when the YAML used the object form.
	Legacy bool
}

// legacyTransport is the YAML shape of the object form.
type legacyTransport struct {
	Maestro *MaestroTransportConfig `yaml:"maestro,omitempty"`
	Client  string                  `yaml:"client"`
}

// NamedTransport returns a reference to the named deployment route.
func NamedTransport(name string) *ResourceTransport {
	return &ResourceTransport{Name: name}
}

// UnmarshalYAML accepts a route name or the object form. The object form is
// decoded strictly, matching the task config's KnownFields decoder.
func (t *ResourceTransport) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		return node.Decode(&t.Name)
	case yaml.MappingNode:
		data, err := yaml.Marshal(node)
		if err != nil {
			return err
		}
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		var legacy legacyTransport
		if err := decoder.Decode(&legacy); err != nil {
			return fmt.Errorf("transport: %w", err)
		}
		*t = ResourceTransport{Maestro: legacy.Maestro, Name: legacy.Client, Legacy: true}
		return nil
	default:
		return fmt.Errorf("line %d: transport must be a transport name", node.Line)
	}
}

// MarshalYAML writes the form the transport was loaded from.
func (t ResourceTransport) MarshalYAML() (interface{}, error) {
	if t.Legacy {
		return legacyTransport{Maestro: t.Maestro, Client: t.Name}, nil
	}
	return t.Name, nil
}

// UsesMaestro reports whether any resource uses the maestro transport client.
func UsesMaestro(resources []Resource) bool {
	for i := range resources {
		if resources[i].IsMaestroTransport() {
			return true
		}
	}
	return false
}

// validateLegacyTransport checks a resource that uses the object form. It
// returns maestro=true when the resource selects Maestro, which needs no
// further named-route validation.
func validateLegacyTransport(
	adapter *AdapterConfig, task *AdapterTaskConfig, resource Resource, path string, vars map[string]bool,
) (maestro bool, err error) {
	transport := resource.Transport
	path += "." + FieldTransport
	// TODO(HYPERFLEET-1443): the object form is accepted only without schema_version
	// "2.0". Keep accepting unversioned v1 configs (no v1 diagnostic error) until the
	// cutover lands, or every deployed adapter config fails at startup.
	if task.SchemaVersion == schemaVersionV2 {
		return false, fmt.Errorf("%s: the object form is not supported with schema_version %q; "+
			"name a transport instead", path, schemaVersionV2)
	}
	switch transport.Name {
	case TransportClientKubernetes:
		if transport.Maestro != nil {
			return false, fmt.Errorf("%s.%s is only valid with %s %q",
				path, FieldMaestro, FieldClient, TransportClientMaestro)
		}
		return false, nil
	case TransportClientMaestro:
	default:
		return false, fmt.Errorf("%s.%s must be %q or %q, got %q",
			path, FieldClient, TransportClientKubernetes, TransportClientMaestro, transport.Name)
	}
	if transport.Maestro == nil {
		return false, fmt.Errorf("%s.%s is required for %s %q", path, FieldMaestro, FieldClient, TransportClientMaestro)
	}
	if adapter.Clients.Maestro == nil {
		return false, fmt.Errorf("%s selects %q, but clients.maestro is not configured", path, TransportClientMaestro)
	}
	for _, match := range templateVarRegex.FindAllStringSubmatch(transport.Maestro.TargetCluster, -1) {
		if !isVariableDefinedIn(match[1], vars) {
			return false, fmt.Errorf("%s.%s.%s uses undefined variable %q",
				path, FieldMaestro, FieldTargetCluster, match[1])
		}
	}
	return true, nil
}
