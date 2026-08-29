package config

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config representa a configuração da aplicação
type Config struct {
	Server            ServerConfig            `mapstructure:"server"`
	Services          ServicesConfig          `mapstructure:"services"`
	RabbitMQ          RabbitMQConfig          `mapstructure:"rabbitmq"`
	JWT               JWTConfig               `mapstructure:"jwt"`
	Metrics           MetricsConfig           `mapstructure:"metrics"`
	Log               LogConfig               `mapstructure:"log"`
	Redis             RedisConfig             `mapstructure:"redis"`
	RateLimit         RateLimitConfig         `mapstructure:"rate_limit"`
	ConnectionsHealth ConnectionsHealthConfig `mapstructure:"connections_health"`
	Env               string                  `mapstructure:"env"`
}

// RedisConfig configurações de conexão com Redis
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// RateLimitConfig configurações gerais de Rate Limit
type RateLimitConfig struct {
	Enabled bool                 `mapstructure:"enabled"`
	Rules   RateLimitRulesConfig `mapstructure:"rules"`
}

// RateLimitRulesConfig mapeamento das regras específicas
type RateLimitRulesConfig struct {
	RegisterInit      RateLimitRule `mapstructure:"register_init"`
	RegisterResend    RateLimitRule `mapstructure:"register_resend"`
	RegisterConfirm   RateLimitRule `mapstructure:"register_confirm"`
	Consents          RateLimitRule `mapstructure:"consents"`
	UsersMe           RateLimitRule `mapstructure:"users_me"`
	ConnectionsHealth RateLimitRule `mapstructure:"connections_health"`
	Default           RateLimitRule `mapstructure:"default"`
}

// ConnectionsHealthConfig snapshot autenticado da tela Conexões.
type ConnectionsHealthConfig struct {
	SnapshotTTL  time.Duration     `mapstructure:"snapshot_ttl"`
	LockTTL      time.Duration     `mapstructure:"lock_ttl"`
	ProbeTimeout time.Duration     `mapstructure:"probe_timeout"`
	URLs         map[string]string `mapstructure:"urls"`
}

// RateLimitRule define o limite e a janela de tempo
type RateLimitRule struct {
	Limit  int           `mapstructure:"limit"`
	Window time.Duration `mapstructure:"window"`
}

// ServerConfig configurações do servidor HTTP
type ServerConfig struct {
	Port           string        `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
}

// ServicesConfig configurações dos microserviços
type ServicesConfig struct {
	Auth          ServiceConfig `mapstructure:"auth"`
	User          ServiceConfig `mapstructure:"user"`
	Company       ServiceConfig `mapstructure:"company"`
	Communication ServiceConfig `mapstructure:"communication"`
	UserConsents  ServiceConfig `mapstructure:"user_consents"`
	UserProfile   ServiceConfig `mapstructure:"user_profile"`
}

// ServiceConfig configurações de um microserviço
type ServiceConfig struct {
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Retries int           `mapstructure:"retries"`
}

// RabbitMQConfig configurações do RabbitMQ
type RabbitMQConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	User       string `mapstructure:"user"`
	Password   string `mapstructure:"password"`
	VHost      string `mapstructure:"vhost"`
	Exchange   string            `mapstructure:"exchange"`
	RoutingKey string            `mapstructure:"routing_key"`
	Durable    bool              `mapstructure:"durable"`
	AutoDelete bool              `mapstructure:"auto_delete"`
	Audit      AuditBrokerConfig `mapstructure:"audit"`
}

type AuditBrokerConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Exchange   string `mapstructure:"exchange"`
	RoutingKey string `mapstructure:"routing_key"`
	Durable    bool   `mapstructure:"durable"`
}

// JWTConfig configurações JWT
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Issuer     string `mapstructure:"issuer"`
	Audience   string `mapstructure:"audience"`
	JWKSURL    string `mapstructure:"jwks_url"`
	SigningKey string `mapstructure:"signing_key"`
}

// MetricsConfig configurações de métricas
type MetricsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	ScrapePath string `mapstructure:"scrape_path"`
	Port       string `mapstructure:"port"`
}

// LogConfig configurações de log
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load carrega a configuração
func Load() (*Config, error) {
	// Configurações padrão
	setDefaults()

	// Configurações de ambiente
	viper.SetEnvPrefix("BFF_CORE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Determina o ambiente
	env := os.Getenv("BFF_CORE_ENV")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "" {
		env = viper.GetString("env")
	}
	if env == "" {
		env = "local"
	}

	// Carrega arquivo de configuração base application.yml
	viper.SetConfigName("application")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/bff-core")
	_ = viper.ReadInConfig()

	// Carrega e mescla arquivo específico do ambiente (application-{env}.yml)
	viper.SetConfigName("application-" + env)
	_ = viper.MergeInConfig()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults define valores padrão
func setDefaults() {
	// Server
	viper.SetDefault("server.port", "8382")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.request_timeout", "30s")

	// Services
	viper.SetDefault("services.auth.base_url", "http://localhost:8081")
	viper.SetDefault("services.auth.timeout", "5s")
	viper.SetDefault("services.auth.retries", 3)

	viper.SetDefault("services.user.base_url", "http://localhost:8085")
	viper.SetDefault("services.user.timeout", "5s")
	viper.SetDefault("services.user.retries", 3)

	viper.SetDefault("services.company.base_url", "http://localhost:8083")
	viper.SetDefault("services.company.timeout", "5s")
	viper.SetDefault("services.company.retries", 3)

	viper.SetDefault("services.communication.base_url", "http://localhost:8082")
	viper.SetDefault("services.communication.timeout", "5s")
	viper.SetDefault("services.communication.retries", 3)

	viper.SetDefault("services.user_consents.base_url", "http://localhost:8086")
	viper.SetDefault("services.user_consents.timeout", "5s")
	viper.SetDefault("services.user_consents.retries", 3)

	viper.SetDefault("services.user_profile.base_url", "http://localhost:8091")
	viper.SetDefault("services.user_profile.timeout", "5s")
	viper.SetDefault("services.user_profile.retries", 3)

	// RabbitMQ
	viper.SetDefault("rabbitmq.host", "localhost")
	viper.SetDefault("rabbitmq.port", 5672)
	viper.SetDefault("rabbitmq.user", "guest")
	viper.SetDefault("rabbitmq.password", "guest")
	viper.SetDefault("rabbitmq.vhost", "/")
	viper.SetDefault("rabbitmq.exchange", "ms-communication-exchange-local")
	viper.SetDefault("rabbitmq.routing_key", "communication.message.send")
	viper.SetDefault("rabbitmq.durable", true)
	viper.SetDefault("rabbitmq.auto_delete", false)
	viper.SetDefault("rabbitmq.audit.enabled", true)
	viper.SetDefault("rabbitmq.audit.exchange", "srv-audit-exchange-local")
	viper.SetDefault("rabbitmq.audit.routing_key", "audit.event")
	viper.SetDefault("rabbitmq.audit.durable", true)

	// JWT
	viper.SetDefault("jwt.issuer", "keepguard")
	viper.SetDefault("jwt.audience", "keepguard-api")
	viper.SetDefault("rate_limit.rules.users_me.limit", 60)
	viper.SetDefault("rate_limit.rules.users_me.window", "60s")
	viper.SetDefault("rate_limit.rules.connections_health.limit", 20)
	viper.SetDefault("rate_limit.rules.connections_health.window", "60s")
	viper.SetDefault("connections_health.snapshot_ttl", "60s")
	viper.SetDefault("connections_health.lock_ttl", "5s")
	viper.SetDefault("connections_health.probe_timeout", "500ms")

	// Metrics
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.scrape_path", "/metrics")
	viper.SetDefault("metrics.port", "9092")

	// Log
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	// Environment
	viper.SetDefault("env", "dev")
}
