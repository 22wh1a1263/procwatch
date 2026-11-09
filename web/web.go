// Package web serves procwatch data over HTTP as JSON and a small status page.
package web

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/22wh1a1263/procwatch/internal/sysinfo"
)

//go:embed index.html
var static embed.FS

const procRoot = "/proc"

// Handler returns the HTTP handler used by both the server and the Vercel function.
func Handler() http.Handler {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(static, ".")
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/snapshot", func(w http.ResponseWriter, _ *http.Request) {
		s, err := sysinfo.CollectSnapshot(procRoot, "/")
		writeJSON(w, s, err)
	})
	mux.HandleFunc("/api/top", func(w http.ResponseWriter, r *http.Request) {
		n := 10
		if v, err := strconv.Atoi(r.URL.Query().Get("n")); err == nil && v > 0 && v <= 50 {
			n = v
		}
		p, err := sysinfo.TopProcesses(procRoot, n)
		writeJSON(w, p, err)
	})
	mux.HandleFunc("/api/ports", func(w http.ResponseWriter, _ *http.Request) {
		p, err := sysinfo.ListeningPorts(procRoot)
		writeJSON(w, p, err)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, v any, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if v == nil {
		v = []struct{}{}
	}
	json.NewEncoder(w).Encode(v)
}
