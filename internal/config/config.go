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
		// Cost to hash the password . Must be grater than 3.
		PasswordHashCost int `yaml:"password_hash_cost" env-default:"14"`
		// Path to migrations.
		Migrations string `yaml:"migrations_path"`
	}
	// Config for HTTP server.
	HTTPServer struct {
		// The server startup address.
		Address string `yaml:"run_address" env:"RUN_ADDRESS" env-default:"127.0.0.1:8080"`
		// Read header timeout.
		Timeout time.Duration `yaml:"timeout" env-default:"5s"`
		// Idle timeout.
		IdleTimeout time.Duration `yaml:"idle_timeout" end-default:"60s"`
		// Shutdown timeout.
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"30s"`
	}
	// Config for application's logger.
	Logger struct {
		// Path to store log files.
		Path string `ymal:"log_path" env:"LOG_PATH"`
		// Application logging level.
		Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
		// Log files details.
		MaxSizeMB  int `yaml:"max_size_mb"`
		MaxBackups int `yaml:"max_backups"`
		MaxAgeDays int `yaml:"max_age_days"`
	}
	// Config for JWT.
	JWT struct {
		// JWT signing key.
		SigningKey string `yaml:"signing_key" env:"JWT_SIGNING_KEY"`
		// JWT expiration.
		Expiration time.Duration `yaml:"expiration" env:"JWT_EXPIRATION" env-default:"24h"`
	}
)

// Order of loading configuration:
// 1. YAML file
// 2. Flags
// 3. Environment variables

// Load returns an application configuration which is populated
// from the given configuration file, environment variables and flags.
func MustLoad() *Config {
	// Configuration yaml file path.
	configPath := flag.String("config", "./config/local.yml", "path to the config file")
	flag.Parse()

	// Check if file exists.
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %v", err)
	}

	var cfg Config

	// Load from YAML cfg file.
	file, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("failed to open config file: %v", err)
	}
	if err = cleanenv.ParseYAML(file, &cfg); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}

	// Read given flags. If not provided use file values.
	flag.StringVar(&cfg.HTTPServer.Address, "a", cfg.HTTPServer.Address, "server startup address")
	flag.StringVar(&cfg.DSN, "d", cfg.DSN, "server data source name")
	flag.StringVar(&cfg.AccrualAddr, "r", cfg.AccrualAddr, "server address of the accural system")
	flag.Parse()

	// Read environment variables.
	if err = cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read environment variables: %v", err)
	}

	return &cfg
}
