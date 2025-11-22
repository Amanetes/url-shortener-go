package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App    AppConfig
	Server ServerConfig
	Db     DbConfig
	Redis  RedisConfig
}

type AppConfig struct {
	Name string `envconfig:"APP_NAME"`
	Env  string `envconfig:"APP_ENV"  default:"development"`
}

type ServerConfig struct {
	Port                string        `envconfig:"SERVER_PORT"`
	Host                string        `envconfig:"SERVER_HOST"`
	ShutdownGracePeriod time.Duration `envconfig:"SHUTDOWN_GRACE_PERIOD_SECONDS" default:"15s"`
}

type DbConfig struct {
	Host     string `envconfig:"DB_HOST"`
	Port     string `envconfig:"DB_PORT"`
	User     string `envconfig:"DB_USER"`
	Password string `envconfig:"DB_PASSWORD"`
	DbName   string `envconfig:"DB_NAME"`
}

type RedisConfig struct {
	DB       int    `envconfig:"REDIS_DB"       default:"0"`
	Host     string `envconfig:"REDIS_HOST"`
	Port     string `envconfig:"REDIS_PORT"`
	Password string `envconfig:"REDIS_PASSWORD"`
}

var (
	cfg     *Config
	once    sync.Once
	loadErr error
)

func Get() (*Config, error) {
	once.Do(func() {
		cfg, loadErr = loadConfig()
	})

	return cfg, loadErr
}

func loadConfig() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	if env == "development" {
		_ = godotenv.Load()
	}

	var c Config

	if err := envconfig.Process("", &c); err != nil {
		return nil, fmt.Errorf("failed to process environment: %w", err)
	}

	return &c, nil
}

func (c *DbConfig) Dsn() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DbName)
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
