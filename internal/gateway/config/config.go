package config

import "os"

type Config struct {
	HTTPPort string
	AccountServiceAddr string
	PaymentServiceAddr string
}

func Load() (*Config, error) {
	httpPort := getEnv("GATEWAY_HTTP_PORT", "8080")
	accountServiceAddr := getEnv("ACCOUNT_SERVICE_ADDR", "localhost:50051")
	paymentServiceAddr := getEnv("PAYMENT_SERVICE_ADDR", "localhost:50052")

	return &Config{
		HTTPPort:           httpPort,
		AccountServiceAddr: accountServiceAddr,
		PaymentServiceAddr: paymentServiceAddr,
	}, nil
}

func getEnv(key, defaultVault string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVault
}