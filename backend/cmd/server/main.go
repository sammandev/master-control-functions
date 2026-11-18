// Command server runs the Web Program Functions Database app: a Fiber HTTP
// server serving the REST API and the built SvelteKit SPA, backed by PostgreSQL.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webprogfndb/internal/api"
	"webprogfndb/internal/db"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	webDir := flag.String("web", "frontend/build", "directory containing the built SvelteKit static bundle (repo-root relative)")
	drain := flag.Duration("shutdown-timeout", 45*time.Second, "max time to drain in-flight requests on SIGINT/SIGTERM")
	flag.Parse()

	d, err := db.OpenPostgres(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer d.Close()
	if err := db.MigratePostgres(d); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	v, err := db.CurrentVersion(d)
	if err != nil {
		log.Fatalf("read schema version: %v", err)
	}
	log.Printf("PostgreSQL database at schema version %d", v)

	app := api.New(d)
	if api.RegisterStatic(app, *webDir) {
		log.Printf("serving SPA bundle from %s", *webDir)
	} else {
		log.Printf("no SPA bundle at %s; API-only mode", *webDir)
	}

	// Compose allows 60s (stop_grace_period) after SIGTERM; drain within that.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("listening on http://%s", *addr)
	if err := api.Serve(ctx, app, *addr, *drain); err != nil {
		log.Fatalf("server: %v", err)
	}
	log.Printf("server stopped")
}
