package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/22wh1a1263/procwatch/web"
)

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fs.String("addr", ":"+port, "listen address (defaults to $PORT, or 8080)")
	fs.Parse(args)
	srv := &http.Server{
		Addr:              *addr,
		Handler:           web.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Println("listening on", *addr)
	return srv.ListenAndServe()
}
