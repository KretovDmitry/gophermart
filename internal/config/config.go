package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config represents an application configuration.
	Config struct {
		// The data source name (DSN) for connecting to the database.
		DSN string `yaml:"dsn" env:"DATABASE_URI"`
		// The address of the accural system server.
		AccrualAddr string `yaml:"accrual_addr" env:"ACCRUAL_SYSTEM_ADDRESS"`
		// Subconfigs.
		HTTPServer HTTPServer `yaml:"http_server"`
		JWT        JWT        `yaml:"jwt"`
		Logger     Logger     `yaml:"logger"`
		// Cost of the password to hash. Must be grater than 3.
		PasswordHashCost int `yaml:"password_hash_cost" env-default:"14"`
	}
	// Config for HTTP server.
	HTTPServer struct {
		// The server startup address.
		Address string `yaml:"run_address" env:"RUN_ADDRESS" env-default:"127.0.0.1:8080"`
		// Read Header Timeout in seconds.
		Timeout time.Duration `yaml:"timeout" env-default:"5s"`
		// Idle timeoutin in seconds.
		IdleTimeout time.Duration `yaml:"idle_timeout" end-default:"60s"`
		// Shutdown timeout in seconds.
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"30s"`
	}
	// Config for application's logger.
	Logger struct {
		// Path to store log files.
		Path string `ymal:"path" env:"LOG_PATH"`
		// Application logging level.
		Level string `yaml:"level" env:"LOG_LEVEL"`
		// Log files details.
		MaxSizeMB  int `yaml:"max_size_mb"`
		MaxBackups int `yaml:"max_backups"`
		MaxAgeDays int `yaml:"max_age_days"`
	}
	// Config for JWT.
	JWT struct {
		// JWT signing key.
		SigningKey string `yaml:"signing_key" env:"JWT_SIGNING_KEY"`
		// JWT expiration in hours.
		Expiration time.Duration `yaml:"expiration" env:"JWT_EXPIRATION" env-default:"24h"`
	}
)

// Load returns an application configuration which is populated
// from the given configuration file, environment variables and flags.
func MustLoad() *Config {
	// Configuration yaml file path.
	configPath := flag.String("config", "./config/local.yml", "path to the config file")
	flag.Parse()

	// Check if file exists.
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", *configPath)
	}

	var cfg Config

	// Load from YAML cfg file.
	bytes, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("failed to open config file: %s", *configPath)
	}
	if err = cleanenv.ParseYAML(bytes, &cfg); err != nil {
		log.Fatalf("failed to parse config file: %s", *configPath)
	}

	// Read given flags.
	flag.StringVar(&cfg.HTTPServer.Address, "a", "127.0.0.1:8080", "server startup address")
	flag.StringVar(&cfg.DSN, "d", "", "server data source name")
	flag.StringVar(&cfg.AccrualAddr, "r", "", "server address of the accural system")
	flag.Parse()

	// Read environment variables.
	if err = cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read environment variables: %v", err)
	}

	return &cfg
}
