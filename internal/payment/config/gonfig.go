package config

import (
	"fmt"
	"os"
)


type Config struct {
	GRPCPort string
	DBURL string
	AccountServiceAddr string
}

func Load() (*Config, error) {
	port := getEnv("PAYMENT_SERVICE_GRPC_PORT", "50052")
	dbURL := getEnv("PAYMENT_SERVICE_DB_URL", "postgres://postgres:postgresadmin@127.0.0.1:5432/payments_db?sslmode=disable")
	accountAddr := getEnv("ACCOUNT_SERVICE_ADDR", "127.0.0.1:50051")

	if dbURL == "" {
		return  nil, fmt.Errorf("PAYMENT_SERVICE_DB_URL is required")
	}

	return &Config{
		GRPCPort:           port,
		DBURL:              dbURL,
		AccountServiceAddr: accountAddr,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}