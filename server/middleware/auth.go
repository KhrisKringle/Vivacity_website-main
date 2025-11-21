package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

// Define a key for using context to pass the user ID.
type contextKey string

const UserIDKey = contextKey("userID")

// SessionAuthMiddleware checks for a valid user session and adds the UserID to the request context.
func SessionAuth(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the session from the request.
			session, err := store.Get(r, "vivacity-session")
			if err != nil || session.IsNew {
				http.Error(w, "Unauthorized: Please log in.", http.StatusUnauthorized)
				return
			}

			// Get the UserID stored during login. It's a string in the session.
			userID, ok := session.Values["UserID"].(string)
			if !ok || userID == "" {
				http.Error(w, "Unauthorized: Invalid session data.", http.StatusUnauthorized)
				return
			}

			// Add the UserID to the request's context so the next handler can access it.
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext is a helper to safely retrieve the UserID from a request's context.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func GetTeamIDFromContext(ctx context.Context) (int64, bool) {
	teamID, ok := ctx.Value("TeamID").(int64)
	return teamID, ok
}
