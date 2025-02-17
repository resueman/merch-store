include .env

# Путь к бинарникам
LOCAL_BIN:=$(CURDIR)/bin

# Путь к спецификации OpenAPI
OPENAPI_SPEC_PATH = ./api/openapi/$(VERSION)/schema.yaml

# Пути для сгенерированных по openapi компонентов
COMPONENTS_DIR = ./internal/api
SERVER_FILE = $(COMPONENTS_DIR)/$(VERSION)/server.go
TYPES_FILE = $(COMPONENTS_DIR)/$(VERSION)/dto.go

# Пути к интерфейсам и мокам
REPO_INTERFACES_PATH = ./internal/repo/repo.go
TX_MANAGER_INTERFACE_PATH = ./pkg/db/db.go
PASSWORD_MANAGER_INTERFACE_PATH = ./pkg/password/bryptManager.go
MOCKS_DIR = test/mocks/

# Установка линтера
install-lint:
	GOBIN=$(LOCAL_BIN) go install -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4

# Запуск линтера
run-lint:
	GOBIN=$(LOCAL_BIN) golangci-lint run ./... --config .golangci.yml

# Установка зависимостей, необходимых при разработке.
install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1
	GOBIN=$(LOCAL_BIN) go install -mod=mod github.com/pressly/goose/v3/cmd/goose@v3.24.1
	#GOBIN=$(LOCAL_BIN) go install github.com/gojuno/minimock/v3/cmd/minimock@v3.4.5
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@v1.6.0

# Запуск сервера и бд с миграциями в режиме
up:
	docker compose --profile development up -d

# Запуск тестовой базы данных в докере, на ней запускаются интеграционные тесты
up-test-db:
	docker compose --profile test up -d

# Билд и запуск сервера на порту 8080
up-server:
	docker build -t merch_store .
	docker run -p 8080:8080 merch_store

# Запуск только бд с миграциями
up-db:
	docker compose up -d postgres_db default_migrate

# Проверка, docker и docker compose установлены
check-docker-installed:
	docker --version
	docker compose version

# Запуск всех тестов
run-tests:
	go test -v ./...

# Получение актуального тестового покрытия
coverage-total:
	go test -coverprofile=coverage.out ./... > /dev/null
	echo -n "\nTotal coverage: " && go tool cover -func=coverage.out | grep total | awk '{print $$3}'

# Получение тестового покрытия usecase слоя
coverage-usecase:
	go test -coverprofile=usecase_coverage.out ./internal/usecase/...
	go tool cover -func=usecase_coverage.out

# Генерация dto из openapi спецификации
gen-dto:
	mkdir -p $(COMPONENTS_DIR)/$(VERSION)
	$(LOCAL_BIN)/oapi-codegen -generate types -package $(VERSION) -o $(TYPES_FILE) $(OPENAPI_SPEC_PATH)

# Удаление сгенерированных dto
rm-dto:
	rm -rf $(COMPONENTS_DIR)/$(VERSION)

# Генерация миграций
gen-migration:
	$(LOCAL_BIN)/goose -dir $(DST) create init sql

# Генерация моков для usecase слоя
gen-usecase-mocks:
	mkdir -p $(MOCKS_DIR)
	$(LOCAL_BIN)/mockgen -source=$(TX_MANAGER_INTERFACE_PATH) -destination=$(MOCKS_DIR)/tx_manager_mock.go -package=mocks
	$(LOCAL_BIN)/mockgen -source=$(REPO_INTERFACES_PATH) -destination=$(MOCKS_DIR)/repo_mock.go -package=mocks
	$(LOCAL_BIN)/mockgen -source=$(PASSWORD_MANAGER_INTERFACE_PATH) -destination=$(MOCKS_DIR)/password_manager_mock.go -package=mocks

# Помощь по доступным таргетам
help:
	@echo "Доступные таргеты:"
	@echo "  install-lint         - Установка линтера"
	@echo "  run-lint             - Запуск линтера"
	@echo "  install-deps         - Установка зависимостей"
	@echo "  up                   - Запуск сервера и БД"
	@echo "  up-test-db           - Запуск тестовой БД"
	@echo "  up-server            - Билд и запуск сервера"
	@echo "  up-db                - Запуск только БД"
	@echo "  check-docker-installed - Проверка установки Docker"
	@echo "  run-tests            - Запуск всех тестов"
	@echo "  coverage-total       - Получение общего тестового покрытия"
	@echo "  coverage-usecase     - Получение покрытия usecase слоя"
	@echo "  gen-dto              - Генерация dto из openapi спецификации"
	@echo "  rm-dto               - Удаление сгенерированных dto"
	@echo "  gen-migration        - Генерация миграций"
	@echo "  gen-usecase-mocks    - Генерация моков для usecase слоя"
