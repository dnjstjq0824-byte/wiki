package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dnjstjqppk-dev/treewiki/services/gateway"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := gateway.NewRouter()

	log.Printf("[Treewiki Gateway] Listening on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server terminated: %v", err)
	}
}

