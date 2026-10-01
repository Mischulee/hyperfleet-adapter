package transportclient

import "context"

type resourceNameKey struct{}

// WithResourceName returns a copy of ctx naming the task config resource that
// the following transport calls are made for. Two resources can render the same
// object for different routes, so the dry-run recorder uses this name, not the
// object identity, to tie its records to a resource.
func WithResourceName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, resourceNameKey{}, name)
}

// ResourceNameFromContext returns the name set by WithResourceName, or "".
func ResourceNameFromContext(ctx context.Context) string {
	if name, ok := ctx.Value(resourceNameKey{}).(string); ok {
		return name
	}
	return ""
}
