.PHONY: build build-api build-agent

build: build-api build-agent

build-api:
	cd api && GOOS=linux GOARCH=amd64 go build -o ../dist/vm-monitor-api ./cmd/api

build-agent:
	cd agent && GOOS=linux GOARCH=amd64 go build -o ../dist/vm-monitor-agent ./cmd/agent
