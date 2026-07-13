.PHONY: update clean build build-all run package deploy test authors dist

gosec:
	$(info Run gosec)
	gosec -color -nosec -tests ./...

staticcheck:
	$(info Run staticcheck)
	staticcheck ./...


vulncheck:
	$(info Run vulncheck)
	govulncheck -show verbose ./...

start:
	$(info Run!)
	docker compose -f docker-compose.yaml up -d --remove-orphans

stop:
	$(info stopping VC)
	docker compose -f docker-compose.yaml rm -s -f

clean:
	$(info Cleaning up)
	docker volume rm ip_service_kv_data

restart: stop start

ifndef VERSION
VERSION := latest
endif

DOCKER_TAG_IP_SERVICE 		:= docker.sunet.se/ip_service:$(VERSION)

docker-build-ip_service:
	$(info Docker Building ip_service with tag: $(VERSION))
	docker build \
		--build-arg GIT_COMMIT=$$(git rev-list -1 HEAD) \
		--build-arg GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD) \
		--tag $(DOCKER_TAG_IP_SERVICE) .

dev_turnover: stop clean docker-build-ip_service start
	$(info Run in dev mode)

docker-push:
	$(info Docker Pushing ip_service with tag: $(VERSION))
	docker push $(DOCKER_TAG_IP_SERVICE)

build-tester:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/tester_ip -ldflags "-w -s --extldflags '-static'" ./cmd/tester/main.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/ip_service -ldflags "-w -s --extldflags '-static'" ./cmd/ip_service/main.go

swagger: swagger-ip_service swagger-fmt

swagger-fmt:
	swag fmt

swagger-ip_service:
	swag init -d internal/apiv1 -g client.go --output docs/ --parseDependency --packageName docs

install-container-tools:
	$(info Install from go)
	go install github.com/swaggo/swag/cmd/swag@latest

diagram:
	plantuml docs/diagrams/*.puml

vscode:
	$(info Install APT packages)
	sudo apt-get update && sudo apt-get install -y \
		protobuf-compiler \
		netcat-openbsd \
		plantuml
	$(info Install go packages)
	go install github.com/swaggo/swag/cmd/swag@latest && \
	go install golang.org/x/tools/cmd/deadcode@latest && \
	go install github.com/securego/gosec/v2/cmd/gosec@latest && \
	go install honnef.co/go/tools/cmd/staticcheck@latest && \
	go install golang.org/x/vuln/cmd/govulncheck@latest && \
	go install golang.org/x/tools/gopls@latest