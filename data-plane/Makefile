run_server:
	go run ./cmd/server/main.go

run_cli:
	go run ./cmd/cli/main.go $(ARGS)

build_cli:
	go build -o nexus-cli ./cmd/cli/main.go

tests:
	go test -v ./...