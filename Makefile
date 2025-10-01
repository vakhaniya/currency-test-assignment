setup:
	go install github.com/swaggo/swag/cmd/swag@latest
	go mod download

test:
	go test -race -v ./...

run:
	go run cmd/main.go

build:
	go build -o bin/app ./cmd/main.go

build-ci:
	CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

build-swagger:
	swag init -g cmd/main.go

