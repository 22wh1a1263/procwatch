// Package handler is the Vercel serverless entry point. All routes are
// rewritten to it (see vercel.json) and served by the shared web.Handler.
package handler

import (
	"net/http"

	"github.com/22wh1a1263/procwatch/web"
)

var h = web.Handler()

// Handler serves the status page and JSON API.
func Handler(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r)
}
