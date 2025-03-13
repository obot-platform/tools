package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adrg/xdg"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/glebarez/sqlite"
	"github.com/gptscript-ai/gptscript-helper-sqlite/pkg/common"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	s, err := NewSqlite(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error creating sqlite: %v\n", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(s, credentials.ActionStore, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(s, credentials.ActionGet, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	mux.HandleFunc("/erase", func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(s, credentials.ActionErase, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(s, credentials.ActionList, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	if err := http.ListenAndServe("127.0.0.1:"+port, mux); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start server: %v", err)
	}
}

func NewSqlite(ctx context.Context) (common.Database, error) {
	var (
		dbPath string
		err    error
	)
	if os.Getenv("GPTSCRIPT_SQLITE_FILE") != "" {
		dbPath = os.Getenv("GPTSCRIPT_SQLITE_FILE")
	} else {
		dbPath, err = xdg.ConfigFile("gptscript/credentials.db")
		if err != nil {
			return common.Database{}, fmt.Errorf("failed to get credentials db path: %w", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
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
