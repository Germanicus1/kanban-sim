package middleware

import (
	"net/http"
	"os"
	"strings"
)

// APIKeyAuth is a simple middleware that enforces a static API key.
// It expects: Authorization: Bearer <GAME_API_KEY>
func APIKeyAuth(next http.Handler) http.Handler {
	key := os.Getenv("API_KEY")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1) Grab "Authorization" header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		// 2) Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid Authorization format", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		// 3) Compare to environment variable
		if token == "" || token != key {
			http.Error(w, "invalid API key", http.StatusForbidden)
			return
		}

		// 4) (Optional) Check Origin or Referrer if you want an extra layer—
		//    e.g., r.Header.Get("Origin") == "https://yourgame.com"
		//    For now, we’ll skip that. If you want, uncomment below:
		//
		// origin := r.Header.Get("Origin")
		// if origin != "https://yourgame.com" {
		//     http.Error(w, "forbidden origin", http.StatusForbidden)
		//     return
		// }

		// 5) All good → call the next handler
		next.ServeHTTP(w, r)
	})
}
