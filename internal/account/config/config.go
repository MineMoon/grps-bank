package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCPort string
	DBURL string
}

func Load() (*Config, error) {
	port := getEnv("ACCOUNT_SERVICE_GRPC_PORT", "50051")
	dbURL := getEnv("ACCOUNT_SERVICE_DB_URL", "postgres://postgres:postgresadmin@127.0.0.1:5432/accounts_db?sslmode=disable")

	if dbURL == "" {
		return nil, fmt.Errorf("переменная ACCOUNT_SERVICE_DB_URL обязательна")
	}

	return &Config{
		GRPCPort: port,
		DBURL:    dbURL,
	}, nil
}

func getEnv(key, defaulValue string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaulValue
}