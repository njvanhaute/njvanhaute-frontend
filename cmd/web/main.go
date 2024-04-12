package main

import (
	"context"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/lib/pq"
)

type application struct {
	logger          *slog.Logger
	sessionManager  *scs.SessionManager
	templateCache   map[string]*template.Template
	formDecoder     *form.Decoder
	httpClient      *http.Client
	backendHostname string
}

type config struct {
	addr string
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	backendHostname   string
	apiMaxRequestTime time.Duration
}

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":4200", "HTTP network address")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.StringVar(&cfg.backendHostname, "backend-hostname", "http://localhost:4000", "Backend API hostname")

	flag.DurationVar(&cfg.apiMaxRequestTime, "api-max-request-time", 10*time.Second, "Backend API max time to wait for repsonse")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)

	formDecoder := form.NewDecoder()

	httpClient := &http.Client{
		Timeout: cfg.apiMaxRequestTime,
	}

	app := &application{
		logger:          logger,
		sessionManager:  sessionManager,
		templateCache:   templateCache,
		formDecoder:     formDecoder,
		httpClient:      httpClient,
		backendHostname: cfg.backendHostname,
	}

	logger.Info("starting server", "addr", cfg.addr)

	err = http.ListenAndServe(cfg.addr, app.routes())
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
