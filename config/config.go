package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Log       LogConfig
	Metrics   MetricsConfig
	Tracing   TracingConfig
	RateLimit RateLimitConfig
}

type RateLimitConfig struct {
	MaxFailures int           // failed login attempts before IP is jailed
	JailMinutes time.Duration // how long the IP stays jailed
}

type TracingConfig struct {
	Endpoint string
	Sampler  float64
}

type MetricsConfig struct {
	Port string
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	viper.SetDefault("JAIL_MAX_FAILURES", 5)
	viper.SetDefault("JAIL_DURATION_MINUTES", 15)

	viper.AutomaticEnv()

	accessExpiry, err := time.ParseDuration(viper.GetString("JWT_ACCESS_TOKEN_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TOKEN_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(viper.GetString("JWT_REFRESH_TOKEN_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TOKEN_EXPIRY: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: viper.GetString("SERVER_HOST"),
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			DBName:   viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			Secret:             viper.GetString("JWT_SECRET"),
			AccessTokenExpiry:  accessExpiry,
			RefreshTokenExpiry: refreshExpiry,
		},
		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
		Metrics: MetricsConfig{
			Port: viper.GetString("METRICS_PORT"),
		},
		Tracing: TracingConfig{
			Endpoint: viper.GetString("TRACING_ENDPOINT"),
			Sampler:  viper.GetFloat64("TRACING_SAMPLER"),
		},
		RateLimit: RateLimitConfig{
			MaxFailures: viper.GetInt("JAIL_MAX_FAILURES"),
			JailMinutes: time.Duration(viper.GetInt("JAIL_DURATION_MINUTES")) * time.Minute,
		},
	}
	return cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}
