.PHONY: run build test clean dev

run:
	go run cmd/main.go

build:
	go build -o bin/auth-service cmd/main.go

dev:
	air

test:
	go test -v ./...

clean:
	rm -rf bin/

deps:
	go mod download
	go mod tidy

docker-build:
	docker build -t auth-service:latest .

docker-run:
	docker run --env-file .env -p 8080:8080 auth-service:latest