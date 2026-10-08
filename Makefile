build:
	go build -o agentdock ./cmd/agentdock

test:
	go test ./...

vet:
	go vet ./...

images:
	docker build -t agentdock/codex:0.161.2 -f docker/codex/Dockerfile .
	docker build -t agentdock/pi:0.73.1 -f docker/pi/Dockerfile .
