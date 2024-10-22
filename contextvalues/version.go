package contextvalues

import "context"

// Define a custom type for context keys to avoid collisions
type versionKeyType string

var versionKey = versionKeyType("version")

// Add version to context
func AddVersionToContext(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, versionKey, version)
}

// Retrieve version from context
func VersionFromContext(ctx context.Context) string {
	version := ""
	if v, ok := ctx.Value(versionKey).(string); ok {
		version = v
	}
	return version
}
