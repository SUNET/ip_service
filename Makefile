.PHONY: update clean build build-all run package deploy test authors dist deadcode gh-install gh-auth release check_current_branch

gosec:
	$(info Run gosec)
	gosec -color -nosec -tests ./...

staticcheck:
	$(info Run staticcheck)
	staticcheck ./...

deadcode:
	$(info Run deadcode)
	deadcode -test ./...

vulncheck:
	$(info Run vulncheck)
	govulncheck -show verbose ./...

test:
	$(info Run tests)
	go test -race -count=1 ./...

test-cover:
	$(info Run tests with coverage)
	go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

test-cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html

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

NAME                    := ip_service
REGISTRY                := docker.sunet.se
CURRENT_BRANCH          := $(shell git rev-parse --abbrev-ref HEAD)

DOCKER_TAG_IP_SERVICE 		:= $(REGISTRY)/$(NAME):$(VERSION)

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

# ==============================================================================
# Release Management
# ==============================================================================

BUMP                    ?= patch
FORCE                   ?=

check_current_branch:
	$(info Current branch: $(CURRENT_BRANCH))
ifeq ($(CURRENT_BRANCH),main)
	$(info On main branch)
else
ifneq ($(FORCE),true)
	$(error Not on main branch — use FORCE=true to override)
else
	$(warning Not on main branch — continuing because FORCE=true)
endif
endif

release: check_current_branch ## Create and push a git tag (BUMP=major|minor|patch)
	@echo "$(BUMP)" | grep -qE '^(major|minor|patch)$$' || \
		{ echo "Error: BUMP must be major, minor, or patch (got: $(BUMP))"; exit 1; }
	@if [ "$(FORCE)" != "true" ] && ! git diff --quiet HEAD 2>/dev/null; then \
		echo "Error: working tree is dirty — commit or stash changes first (use FORCE=true to override)"; exit 1; \
	fi
	@LATEST=$$(git tag -l "v*" --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | head -n1); \
	if [ -z "$$LATEST" ]; then \
		echo "No existing version tags found, starting at v0.0.0"; \
		LATEST="v0.0.0"; \
	fi; \
	CURRENT=$$(echo "$$LATEST" | sed 's/^v//'); \
	MAJOR=$$(echo "$$CURRENT" | cut -d. -f1); \
	MINOR=$$(echo "$$CURRENT" | cut -d. -f2); \
	PATCH=$$(echo "$$CURRENT" | cut -d. -f3); \
	case "$(BUMP)" in \
		major) MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0 ;; \
		minor) MINOR=$$((MINOR + 1)); PATCH=0 ;; \
		patch) PATCH=$$((PATCH + 1)) ;; \
	esac; \
	NEW_TAG="v$${MAJOR}.$${MINOR}.$${PATCH}"; \
	DOCKER_IMAGE="$(REGISTRY)/$(NAME):$$NEW_TAG"; \
	DOCKER_LATEST="$(REGISTRY)/$(NAME):latest"; \
	echo ""; \
	echo "Bumping $$LATEST -> $$NEW_TAG ($(BUMP))"; \
	echo ""; \
	echo "==> Building Docker image $$DOCKER_IMAGE"; \
	docker build \
		--build-arg GIT_COMMIT=$$(git rev-list -1 HEAD) \
		--build-arg GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD) \
		--tag "$$DOCKER_IMAGE" --tag "$$DOCKER_LATEST" .; \
	echo "==> Pushing $$DOCKER_IMAGE"; \
	docker push "$$DOCKER_IMAGE"; \
	echo "==> Pushing $$DOCKER_LATEST (-> $$NEW_TAG)"; \
	docker push "$$DOCKER_LATEST"; \
	echo ""; \
	echo "==> Tagging git $$NEW_TAG"; \
	git tag -a "$$NEW_TAG" -m "Release $$NEW_TAG"; \
	git push origin "$$NEW_TAG"; \
	echo ""; \
	echo "==> Release $$NEW_TAG created and pushed"; \
	echo ""

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

vscode: gh-install
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

gh-install:
	$(info Install GitHub CLI)
	@if ! command -v gh >/dev/null 2>&1; then \
		sudo mkdir -p -m 755 /etc/apt/keyrings && \
		curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null && \
		sudo chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg && \
		echo "deb [arch=$$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null && \
		sudo apt-get update && \
		sudo apt-get install -y gh; \
	else \
		echo "gh already installed: $$(gh --version | head -1)"; \
	fi

gh-auth: gh-install
	$(info Authenticate GitHub CLI)
	@if gh auth status >/dev/null 2>&1; then \
		gh auth status; \
	else \
		gh auth login; \
	fi