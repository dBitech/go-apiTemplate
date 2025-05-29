// Package config provides application configuration management.
// It handles loading and validation of configuration from environment variables and files.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Tracing  TracingConfig  `mapstructure:"tracing"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

// ServerConfig holds all server related configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout  time.Duration `mapstructure:"idleTimeout"`
}

// DatabaseConfig holds all database related configuration
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslMode"`
}

// LoggingConfig holds all logging related configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// MetricsConfig holds all metrics related configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
}

// TracingConfig holds all tracing related configuration
type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"serviceName"`
}

// AuthConfig holds all authentication related configuration
type AuthConfig struct {
	Enabled            bool          `mapstructure:"enabled"`
	JWTSecret          string        `mapstructure:"jwtSecret"`
	JWTSigningMethod   string        `mapstructure:"jwtSigningMethod"`
	JWTExpirationTime  time.Duration `mapstructure:"jwtExpirationTime"`
	JWTIssuer          string        `mapstructure:"jwtIssuer"`
	OAuth2ClientID     string        `mapstructure:"oauth2ClientID"`
	OAuth2ClientSecret string        `mapstructure:"oauth2ClientSecret"`
	OAuth2RedirectURL  string        `mapstructure:"oauth2RedirectURL"`
	OAuth2AuthURL      string        `mapstructure:"oauth2AuthURL"`
	OAuth2TokenURL     string        `mapstructure:"oauth2TokenURL"`
	OAuth2Scopes       []string      `mapstructure:"oauth2Scopes"`
}

// Load loads the configuration from environment variables, config file, and command line flags
func Load() (*Config, error) {
	// Set default config
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.readTimeout", 10*time.Second)
	viper.SetDefault("server.writeTimeout", 10*time.Second)
	viper.SetDefault("server.idleTimeout", 60*time.Second)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.host", "0.0.0.0")
	viper.SetDefault("metrics.port", 9090)
	viper.SetDefault("tracing.enabled", true)
	viper.SetDefault("tracing.endpoint", "localhost:4317")
	viper.SetDefault("tracing.serviceName", "api-service")
	viper.SetDefault("auth.enabled", true)
	viper.SetDefault("auth.jwtSecret", "your-secret-key-change-me-in-production")
	viper.SetDefault("auth.jwtSigningMethod", "HS256")
	viper.SetDefault("auth.jwtExpirationTime", 24*time.Hour)
	viper.SetDefault("auth.jwtIssuer", "api-template")
	viper.SetDefault("auth.oauth2ClientID", "example-client-id")
	viper.SetDefault("auth.oauth2ClientSecret", "example-client-secret")
	viper.SetDefault("auth.oauth2RedirectURL", "http://localhost:8080/auth/callback")
	viper.SetDefault("auth.oauth2AuthURL", "https://example.com/oauth/authorize")
	viper.SetDefault("auth.oauth2TokenURL", "https://example.com/oauth/token")
	viper.SetDefault("auth.oauth2Scopes", []string{"read", "write"})

	// Environment variables
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/app")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, continue with defaults and env vars
	}

	// Command line flags
	pflag.String("config", "", "Path to config file")
	pflag.String("server.host", viper.GetString("server.host"), "Server host")
	pflag.Int("server.port", viper.GetInt("server.port"), "Server port")
	pflag.String("logging.level", viper.GetString("logging.level"), "Logging level")
	pflag.String("logging.format", viper.GetString("logging.format"), "Logging format (json or text)")
	pflag.Bool("metrics.enabled", viper.GetBool("metrics.enabled"), "Enable Prometheus metrics")
	pflag.Bool("tracing.enabled", viper.GetBool("tracing.enabled"), "Enable OpenTelemetry tracing")
	pflag.Parse()

	// Bind flags to viper
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Check for custom config file specified via flag
	if configFile := viper.GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// String returns a string representation of the config for logging
func (c *Config) String() string {
	// Hide sensitive information
	return fmt.Sprintf(
		"Server: %s:%d, Logging: level=%s format=%s, Metrics: enabled=%t port=%d, Tracing: enabled=%t",
		c.Server.Host,
		c.Server.Port,
		c.Logging.Level,
		c.Logging.Format,
		c.Metrics.Enabled,
		c.Metrics.Port,
		c.Tracing.Enabled,
	)
}
