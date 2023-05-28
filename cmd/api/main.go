package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Hardcoded for now,
const version = "1.0.0"

// Holds app config settings.
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Holds application-wide dependencies.
type application struct {
	config config
	logger *log.Logger
}

func main() {
	cfg, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDb(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Printf("Database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("Starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func parseFlags() (config, error) {
	var cfg config
	dsn, err := getDsn()

	if err != nil || dsn == "" {
		return cfg, errors.New("make sure a valid GREENLIGHT_DB_DSN environment variable is set in a .env file at the root of the project")
	}

	// Cli flags.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", dsn, "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time (duration)")

	flag.Parse()

	return cfg, nil
}

func getDsn() (string, error) {
	envMap, err := godotenv.Read()
	if err != nil {
		return "", err
	}

	return envMap["GREENLIGHT_DB_DSN"], nil
}

func openDb(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set max open and idle connections, and max idle time with the values defined in config.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	// init context with 5 seconds timeout, used as a deadline for the db connection.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Establish connection to db, if it doesn't within the 5 seconds timeout, return error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
