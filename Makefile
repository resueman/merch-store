#include .env

# Путь к бинарникам
LOCAL_BIN:=$(CURDIR)/bin

# Путь к спецификации OpenAPI
OPENAPI_SPEC_PATH = ./api/openapi/v1/schema.yaml

# Пути для сгенерированных по openapi компонентов
COMPONENTS_DIR = ./internal/api
SERVER_FILE = $(COMPONENTS_DIR)/$(VERSION)/server.go
TYPES_FILE = $(COMPONENTS_DIR)/$(VERSION)/dto.go

install-lint:
	GOBIN=$(LOCAL_BIN) go install -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4

run-lint:
	GOBIN=$(LOCAL_BIN) golangci-lint run ./... --config .golangci.yml

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1
	GOBIN=$(LOCAL_BIN) go install -mod=mod github.com/pressly/goose/v3/cmd/goose@v3.24.1

gen-api: 
	$(MAKE) gen-dto VERSION=$(VERSION)
	$(MAKE) gen-server VERSION=$(VERSION)

gen-dto:
	mkdir -p $(COMPONENTS_DIR)/$(VERSION)
	$(LOCAL_BIN)/oapi-codegen -generate types -package $(VERSION) -o $(TYPES_FILE) $(OPENAPI_SPEC_PATH)

gen-server:
	mkdir -p $(COMPONENTS_DIR)/$(VERSION)
	$(LOCAL_BIN)/oapi-codegen -generate server -package $(VERSION) -o $(SERVER_FILE) $(OPENAPI_SPEC_PATH)

clean-api:
	rm -rf $(COMPONENTS_DIR)/$(VERSION)

build:
	GOOS=linux GOARCH=amd64 go build -o merch_store_linux cmd/store/main.go

help:
	@echo "Доступные команды:"
	@echo "make gen-api VERSION=v1 - Генерирует dto и сервер версии v1"
	@echo "make gen-dto VERSION=v1 - Генерирует dto версии v1"
	@echo "make gen-server VERSION=v1 - Генерирует сервер версии v1"
	@echo "make clean-api VERSION=v1 - Очищает весь сгенерированный код версии v1"
