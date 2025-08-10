package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"go.rumenx.com/sixtysix"
	"go.rumenx.com/sixtysix/api"
	"go.rumenx.com/sixtysix/engine"
	"go.rumenx.com/sixtysix/store"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	port := flag.String("port", "8080", "listen port")
	flag.Parse()

	mem := store.NewMemory()
	e := engine.New(mem)
	e.Register(sixtysix.Game{})

	srv := api.New(e)
	addr := ":" + *port
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("go-sixtysix server starting | addr=%s version=%s commit=%s date=%s", addr, version, commit, date)
	start := time.Now()
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server error (uptime=%s): %v", time.Since(start), err)
	}
}
