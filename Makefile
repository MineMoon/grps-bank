ACCOUNTS_DB_URL=postgres://postgres:postgresadmin@127.0.0.1:5432/accounts_db?sslmode=disable
PAYMENTS_DB_URL=postgres://postgres:postgresadmin@127.0.0.1:5432/payments_db?sslmode=disable
COMPOSE_FILE=deployments/docker-compose.yml

.PHONY: proto clean docker-up docker-down docker-reset migrate-up migrate-down

# Генерация Protobuf контрактов
proto:
	protoc --proto_path=api/proto \
	       --go_out=. --go_opt=module=grpc-bank \
	       --go-grpc_out=. --go-grpc_opt=module=grpc-bank \
	       api/proto/account/account.proto \
	       api/proto/payment/payment.proto

# Управление базой данных в Docker
docker-up:
	docker compose -f $(COMPOSE_FILE) up -d

docker-down:
	docker compose -f $(COMPOSE_FILE) down

# Полный сброс базы (удаляет том с диска и поднимает начисто)
docker-reset:
	docker compose -f $(COMPOSE_FILE) down -v
	docker compose -f $(COMPOSE_FILE) up -d

# Управление миграциями
migrate-up:
	migrate -path migrations/accounts_db -database "$(ACCOUNTS_DB_URL)" up
	migrate -path migrations/payments_db -database "$(PAYMENTS_DB_URL)" up

migrate-down:
	migrate -path migrations/payments_db -database "$(PAYMENTS_DB_URL)" down -all
	migrate -path migrations/accounts_db -database "$(ACCOUNTS_DB_URL)" down -all

clean:
	rm -rf pkg/api/*