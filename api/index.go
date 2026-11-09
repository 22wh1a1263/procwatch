// Package handler is the Vercel serverless entry point. All routes are
// rewritten to it (see vercel.json) and served by the shared web.Handler.
package handler

import (
	"net/http"
	"os"

	"github.com/22wh1a1263/procwatch/web"
)

var h = func() http.Handler {
	if os.Getenv("VERCEL") != "" {
		return web.HandlerWithDisk("/tmp") // root filesystem is read-only on Vercel
	}
	return web.Handler()
}()

// Handler serves the status page and JSON API.
func Handler(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r)
}
