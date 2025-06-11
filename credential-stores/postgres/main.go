package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/gptscript-ai/gptscript-helper-sqlite/pkg/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var tokenFromEnv = os.Getenv("GPTSCRIPT_DAEMON_TOKEN")

func main() {
	p, err := NewPostgres(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error creating postgres: %v\n", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /store", authenticatedHandler(func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(p, credentials.ActionStore, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}))
	mux.HandleFunc("POST /get", authenticatedHandler(func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(p, credentials.ActionGet, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}))
	mux.HandleFunc("POST /erase", authenticatedHandler(func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(p, credentials.ActionErase, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}))
	mux.HandleFunc("POST /list", authenticatedHandler(func(w http.ResponseWriter, r *http.Request) {
		// Check for context in the body.
		body, err := io.ReadAll(r.Body)
		if err == nil && len(body) > 0 {
			var contexts []string
			if err := json.Unmarshal(body, &contexts); err != nil {
				http.Error(w, "invalid request body: body must be a JSON array of credential context strings", http.StatusBadRequest)
				return
			}

			// Use our custom ListWithContexts handler rather than relying on Docker's libraries.
			creds, err := p.ListWithContexts(contexts)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(creds)
		} else {
			if err := credentials.HandleCommand(p, credentials.ActionList, r.Body, w); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}
	}))
	// Leave this one unauthenticated so that the health check works.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe("127.0.0.1:"+port, mux); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start server: %v", err)
	}
}

func NewPostgres(ctx context.Context) (common.Database, error) {
	dsn := os.Getenv("GPTSCRIPT_POSTGRES_DSN")
	if dsn == "" {
		return common.Database{}, fmt.Errorf("missing GPTSCRIPT_POSTGRES_DSN")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return common.Database{}, fmt.Errorf("failed to open database: %w", err)
	}

	return common.NewDatabase(ctx, db)
}

func authenticatedHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authenticate(r.Header) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func authenticate(headers http.Header) bool {
	return headers.Get("X-GPTScript-Daemon-Token") == tokenFromEnv
}
