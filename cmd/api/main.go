package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"greenlight.ricci2511.dev/internal/data"
	"greenlight.ricci2511.dev/internal/jsonlog"
	"greenlight.ricci2511.dev/internal/mailer"
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
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

// Holds application-wide dependencies.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	cfg, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDb(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	// Set basic application metrics.
	setExpVars(db)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func parseFlags() (config, error) {
	var cfg config
	env, err := getEnvVars()

	if err != nil {
		return cfg, fmt.Errorf("failed to read environment variables: %s", err.Error())
	}

	// Server settings.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Database settings.
	flag.StringVar(&cfg.db.dsn, "db-dsn", env["GREENLIGHT_DB_DSN"], "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time (duration)")

	// Rate limiter settings.
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// SMTP settings.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP server hostname")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP server port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", env["SMTP_USERNAME"], "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", env["SMTP_PASSWORD"], "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.ricci2511.dev>", "SMTP sender")

	// CORS settings.
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	return cfg, nil
}

func getEnvVars() (map[string]string, error) {
	envMap, err := godotenv.Read()
	if err != nil {
		return nil, err
	}

	return envMap, nil
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

func setExpVars(db *sql.DB) {
	expvar.NewString("version").Set(version)

	// Number of active goroutines.
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// Database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	// Current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
}
