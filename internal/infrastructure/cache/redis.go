package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisConfig configurações para conexão com Redis
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// NewRedisClient cria uma nova conexão de cliente com o Redis
func NewRedisClient(cfg RedisConfig, logger *zap.Logger) (*redis.Client, error) {
	host := cfg.Host
	if envHost := os.Getenv("BFF_CORE_REDIS_HOST"); envHost != "" {
		host = envHost
	} else if envHost := os.Getenv("REDIS_HOST"); envHost != "" {
		host = envHost
	}

	if host == "" {
		host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 6379
	}

	addr := fmt.Sprintf("%s:%d", host, cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("Não foi possível conectar ao Redis na inicialização (modo degradado/bypass ativado)",
			zap.String("addr", addr),
			zap.Error(err),
		)
		return client, nil
	}

	logger.Info("Conexão com Redis estabelecida com sucesso", zap.String("addr", addr))
	return client, nil
}
