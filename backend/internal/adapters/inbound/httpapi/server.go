// Package httpapi is the inbound HTTP adapter: the REST API plus the
// embedded PWA static files.
package httpapi

import (
	"crypto/subtle"
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// staticFiles holds the compiled PWA (web/dist). The Docker build copies it
// into ./static before compiling the binary. The "all:" prefix keeps the
// .gitkeep placeholder so the embed pattern always matches something.
//
//go:embed all:static
var staticFiles embed.FS

// NewRouter returns the root HTTP handler: /api/* requires the bearer token,
// everything else serves the PWA so the browser can load it unauthenticated.
func NewRouter(h *Handler, apiToken string) (http.Handler, error) {
	api := http.NewServeMux()
	h.registerRoutes(api)

	static, err := staticHandler()
	if err != nil {
		return nil, err
	}

	root := http.NewServeMux()
	root.Handle("/api/", requireToken(apiToken, api))
	root.Handle("/", static)
	return root, nil
}

// requireToken checks the "Authorization: Bearer <token>" header. With an
// empty token the check is disabled; this is meant for local development only.
func requireToken(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// staticHandler serves the embedded PWA. Paths that are not real files
// (e.g. /tasks/42) fall back to index.html so client-side routing works.
func staticHandler() (http.Handler, error) {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "."
		}
		if _, err := fs.Stat(sub, name); err != nil {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	}), nil
}
