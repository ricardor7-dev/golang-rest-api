package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3"
	"github.com/rs/zerolog"

	"golang-rest-api/db"
	"golang-rest-api/errs"
	"golang-rest-api/logger"
	"golang-rest-api/repositories"
	"golang-rest-api/server"
)

const (
	// log level environment variable name
	loglevelEnv string = "LOG_LEVEL"
	// log error stack environment variable name
	logErrorStackEnv string = "LOG_ERROR_STACK"
	// server host environment variable name
	hostEnv string = "HTTP_SERVER_HOST"
	// server port environment variable name
	portEnv string = "HTTP_SERVER_PORT"
)

func Run(args []string) (err error) {
	const op errs.Op = "cmd/Run"
	var cfg Config

	err = ParseFlags(args, &cfg)
	if err != nil {
		return errs.E(op, err)
	}

	var lvl zerolog.Level
	lvl, err = zerolog.ParseLevel(cfg.Logging.loglvl)
	if err != nil {
		return errs.E(op, err)
	}

	// setup logger with appropriate defaults
	lgr := logger.NewWithGCPHook(os.Stdout, lvl, true) //delete?

	zerolog.SetGlobalLevel(lvl)

	// set global logging time field format to Unix timestamp
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	lgr.Info().Msgf("logging level set to %s", lvl)

	// set global to log errors with stack (or not) based on flag
	logger.WriteErrorStack(cfg.Logging.logErrorStack)
	lgr.Info().Msgf("log error stack global set to %t", cfg.Logging.logErrorStack)

	a := server.New(server.NewMuxRouter(), server.NewDriver(), lgr)

	//DB connection
	dbconn, err := db.Connection(newPostgreSQLDSN(cfg), lgr)
	if err != nil {
		lgr.Fatal().Err(err).Msg("db.Conncection error")
	}

	err = dbconn.AutomigrateTables(lgr)
	if err != nil {
		lgr.Fatal().Err(err).Msg("db.Conncection error")
	}

	// Repositories service
	a.Repositories = server.Repositories{
		AudiobookRepository: repositories.NewAudiobookRepository(dbconn),
		LanguageRepository:  repositories.NewLanguagesRepository(dbconn),
		TagRepository:       repositories.NewTagRepository(dbconn),
	}

	a.Addr = fmt.Sprintf("%s:%s", cfg.HttpServer.Host, cfg.HttpServer.Port)

	lgr.Info().Msgf("serving at %s", a.Addr)

	return a.ListenAndServe()
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags(args []string, config *Config) error {
	const op errs.Op = "cmd/ParseFlags"
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)

	var (
		loglvl        = fs.String("log-level", "info", fmt.Sprintf("sets log level (trace, debug, info, warn, error, fatal, panic, disabled), (also via %s)", loglevelEnv))
		logErrorStack = fs.Bool("log-error-stack", true, fmt.Sprintf("if true, log full error stacktrace, else just log error, (also via %s)", logErrorStackEnv))
		host          = fs.String("http-server-host", "localhost", fmt.Sprintf("listen host for server (also via %s)", hostEnv))
		port          = fs.String("http-server-port", "8080", fmt.Sprintf("listen port for server (also via %s)", portEnv))
		dbhost        = fs.String("db-host", "localhost", fmt.Sprintf("postgresql database host (also via %s)", db.DBHostEnv))
		dbport        = fs.String("db-port", "5432", fmt.Sprintf("postgresql database port (also via %s)", db.DBPortEnv))
		dbname        = fs.String("db-name", "", fmt.Sprintf("postgresql database name (also via %s)", db.DBNameEnv))
		dbuser        = fs.String("db-user", "", fmt.Sprintf("postgresql database user (also via %s)", db.DBUserEnv))
		dbpassword    = fs.String("db-password", "", fmt.Sprintf("postgresql database password (also via %s)", db.DBPasswordEnv))
		dbSSLMode     = fs.String("db-sslmode", "disable", fmt.Sprintf("postgresql database ssl mode (also via %s)", db.DBSSLMode))
	)

	// Parse the command line flags from above
	err := ff.Parse(fs, args[1:], ff.WithEnvVars())
	if err != nil {
		return errs.E(op, err)
	}

	config.Logging.loglvl = *loglvl
	config.Logging.logErrorStack = *logErrorStack

	config.HttpServer.Host = *host
	config.HttpServer.Port = *port

	config.db.host = *dbhost
	config.db.port = *dbport
	config.db.dbname = *dbname
	config.db.userName = *dbuser
	config.db.password = *dbpassword
	config.db.sslMode = *dbSSLMode

	return nil
}

// newPostgreSQLDSN initializes a db.PostgreSQLDSN given a config struct
func newPostgreSQLDSN(cfg Config) db.PostgreSQLDSN {
	return db.PostgreSQLDSN{
		Host:   cfg.db.host,
		Port:   cfg.db.port,
		DBName: cfg.db.dbname,
		//SearchPath: cfg.DBdbsearchpath,
		User:     cfg.db.userName,
		Password: cfg.db.password,
	}
}

type Config struct {
	Logging struct {
		loglvl        string `yaml:"level"`
		logErrorStack bool
	} `yaml:"Logging"`

	HttpServer struct {
		Host string `yaml:"Host"`
		Port string `yaml:"Port"`
	} `yaml:"HttpServer"`

	db struct {
		// Host is the local machine IP Address to bind the HTTP Server to
		host string `yaml:"host"`

		// Port is the local machine TCP Port to bind the HTTP Server to
		port string `yaml:"port"`

		//Application DB Name
		dbname string `yaml:"bdName"`

		//Application userName
		userName string `yaml:"username"`

		//Application password
		password string `yaml:"password"`

		sslMode string `yaml:"sslmode"`
	} `yaml:"db"`
}
