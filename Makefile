test:
	go test -v -race ./...

build:
	go build -v -o ./bin/anti-bruteforce ./cmd/anti-bruteforce/

run: build
	docker compose -f ./deployments/docker-compose.yml up -d
	./bin/anti-bruteforce