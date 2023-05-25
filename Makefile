test:
	go test -v -race -timeout 60s -count 10 ./...

build:
	go build -v -o ./bin/anti-bruteforce ./cmd/anti-bruteforce/
	go build -v -o ./bin/abfcli ./cmd/abfcli/

run: build
	docker compose -f ./deployments/docker-compose.yml up -d
	./bin/anti-bruteforce

lint: 
	golangci-lint run ./...

