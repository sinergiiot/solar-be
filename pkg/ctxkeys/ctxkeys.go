package ctxkeys

// ContextKey is a distinct type for context keys to avoid collisions.
type ContextKey string

const (
	// UserID is the context key for the authenticated user's UUID.
	UserID ContextKey = "auth_user_id"
	// UserRole is the context key for the authenticated user's role.
	UserRole ContextKey = "auth_user_role"
	// IsAPIKey indicates whether the request was authenticated via API key.
	IsAPIKey ContextKey = "auth_is_api_key"
)
